package propagator

import (
	"math"
)

// Probability is a floating point value indicating the chance that an index will be picked.
type Probability = float32

// Priority is a non-negative value indicating the priority with which an index will be picked.
// Lower values precede higher values, 0 being the lowest possible.
type Priority = uint32

var indexFactorySingleton = &indexFactory{
	indices: make(map[packedProbPrio]*index),
}

type indexFactory struct {
	indices map[packedProbPrio]*index
}

// create creates an index from a given probability and priority
func (f *indexFactory) create(probability Probability, priority Priority) *index {
	if math.Abs(float64(probability)) < 1e-10 {
		return nil
	}

	probAndPrio := packPriorityProbability(probability, priority)

	storedIndex, has := f.indices[probAndPrio]
	if !has {
		storedIndex = &index{
			modifiers:   map[constraintId]packedProbPrio{-1: probAndPrio},
			probAndPrio: probAndPrio,
		}
		f.indices[probAndPrio] = storedIndex
	}
	return storedIndex
}

// index stores the probability and priority of a single index in a domain.
type index struct {
	// modifiers contains the probability and priority for each constraint if it applies. Furthermore, it contains the
	// default probability and priority with key -1. These are fixed and cannot be modified.
	modifiers map[constraintId]packedProbPrio

	// Product of probability modifiers in the lower 32 bits and sum of priority modifiers in the higher 32 bits.
	probAndPrio packedProbPrio
}

// adjust adjusts this index according to the given probability and priority for constraintId. It returns the new index
// and whether it was possible at all to adjust this index. Will return a nil index if it is banned (has a probability
// of zero).
func (i *index) adjust(constraint constraintId, probability Probability, priority Priority) (*index, bool) {
	if probability == 0.0 {
		return nil, true
	}

	currentProbability := Probability(1.0)
	currentPriority := Priority(0)
	currentProbAndPrio, has := i.modifiers[constraint]
	if has {
		currentProbability, currentPriority = unpackPriorityProbability(currentProbAndPrio)
	}

	// TODO: think about this: what does it mean if it has/!has the constraint? Is there some optimization here?

	shouldUpdateProbability := probability < currentProbability
	shouldUpdatePriority := priority > currentPriority

	if !shouldUpdateProbability && !shouldUpdatePriority {
		return i, false
	}

	adjustedIndex := &index{
		modifiers: make(map[constraintId]packedProbPrio),
	}

	adjustedProbability := Probability(1.0)
	adjustedPriority := Priority(0)

	prodProbability := Probability(1.0)
	sumPriority := Priority(0)
	for k, modifier := range i.modifiers {
		adjustedIndex.modifiers[k] = modifier
		probability, priority := unpackPriorityProbability(modifier)
		prodProbability = prodProbability * probability
		sumPriority = sumPriority + priority
	}

	if shouldUpdateProbability {
		prodProbability = prodProbability * probability
		if math.Abs(float64(prodProbability)) < 1e-10 {
			return nil, true
		}
		adjustedProbability = probability
	}

	if shouldUpdatePriority {
		sumPriority = sumPriority + priority
		adjustedPriority = priority
	}

	adjustedIndex.modifiers[constraint] = packPriorityProbability(adjustedProbability, adjustedPriority)
	adjustedIndex.probAndPrio = packPriorityProbability(prodProbability, sumPriority)

	return adjustedIndex, true
}

type packedProbPrio int64

func packPriorityProbability(probability Probability, priority Priority) packedProbPrio {
	probabilityBits := math.Float32bits(probability)
	priorityBits := priority

	return packedProbPrio(uint64(probabilityBits)<<32 | uint64(priorityBits))
}

func unpackPriorityProbability(p packedProbPrio) (Probability, Priority) {
	probabilityBits := uint32(p >> 32)
	priorityBits := uint32(p & 0xFFFFFFFF)

	return math.Float32frombits(probabilityBits), priorityBits
}
