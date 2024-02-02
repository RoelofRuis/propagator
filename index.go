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
			probabilityModifiers: map[constraintId]Probability{-1: probability},
			priorityModifiers:    map[constraintId]Priority{-1: priority},
			probAndPrio:          probAndPrio,
		}
		f.indices[probAndPrio] = storedIndex
	}
	return storedIndex
}

// index stores the probability and priority of a single index in a domain.
// The -1 key in both maps serves as the respective base probability and base priority. These are fixed and cannot be
// modified by a constraint.
type index struct {
	probabilityModifiers map[constraintId]Probability
	priorityModifiers    map[constraintId]Priority

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

	currentProbability, has := i.probabilityModifiers[constraint]
	if !has {
		currentProbability = 1.0
	}
	shouldUpdateProbability := probability < currentProbability

	currentPriority, has := i.priorityModifiers[constraint]
	if !has {
		currentPriority = 0
	}
	shouldUpdatePriority := priority > currentPriority

	if !shouldUpdateProbability && !shouldUpdatePriority {
		return i, false
	}

	// FIXME: this is memory consuming
	adjustedIndex := &index{
		probabilityModifiers: i.probabilityModifiers,
		priorityModifiers:    i.priorityModifiers,
	}

	var adjustedProbability Probability
	var adjustedPriority Priority

	if shouldUpdateProbability {
		adjustedIndex.probabilityModifiers = make(map[constraintId]Probability)
		newProbability := float32(1.0)
		for k, prob := range i.probabilityModifiers {
			adjustedIndex.probabilityModifiers[k] = prob
			newProbability = newProbability * prob
		}
		adjustedIndex.probabilityModifiers[constraint] = probability
		adjustedProbability = newProbability * probability
		if math.Abs(float64(adjustedProbability)) < 1e-10 {
			return nil, true
		}
	}

	if shouldUpdatePriority {
		adjustedIndex.priorityModifiers = make(map[constraintId]Priority)
		newPriority := uint32(0)
		for k, prio := range i.priorityModifiers {
			adjustedIndex.priorityModifiers[k] = prio
			newPriority = newPriority + prio
		}
		adjustedIndex.priorityModifiers[constraint] = priority
		adjustedPriority = newPriority + priority
	}

	adjustedIndex.probAndPrio = packPriorityProbability(adjustedProbability, adjustedPriority)

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
