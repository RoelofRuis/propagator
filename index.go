package propagator

import (
	"math"
)

// Probability is a floating point value indicating the chance that an index will be picked.
type Probability = float32

// Priority is a non-negative value indicating the priority with which an index will be picked.
// Lower values precede higher values, 0 being the lowest possible.
type Priority = uint32

var (
	indexFactorySingleton = &indexFactory{
		indices:     make(map[packedProbPrio]*index),
		floatBuffer: make([]byte, 0, 24),
	}
	bannedIndex = &index{
		probabilityModifiers: nil,
		priorityModifiers:    nil,
		probability:          0.0,
		priority:             math.MaxUint32,
		isBanned:             true,
	}
)

type indexFactory struct {
	floatBuffer []byte
	indices     map[packedProbPrio]*index
}

// create creates an index from a given probability and priority
func (f *indexFactory) create(probability Probability, priority Priority) *index {
	if math.Abs(float64(probability)) < 1e-10 {
		return bannedIndex
	}

	packed := packPriorityProbability(probability, priority)

	storedIndex, has := f.indices[packed]
	if !has {
		storedIndex = &index{
			probabilityModifiers: map[constraintId]Probability{-1: probability},
			priorityModifiers:    map[constraintId]Priority{-1: priority},
			probability:          probability,
			priority:             priority,
			isBanned:             false,
		}
		f.indices[packed] = storedIndex
	}
	return storedIndex
}

// index stores the probability and priority of a single index in a domain.
// The -1 key in both maps serves as the respective base probability and base priority. These are fixed and cannot be
// modified by a constraint.
type index struct {
	probabilityModifiers map[constraintId]Probability
	priorityModifiers    map[constraintId]Priority

	// Product of probability modifiers
	probability Probability
	// Sum of priority modifiers
	priority Priority
	// Whether the index is currently banned
	isBanned bool
}

// adjust adjusts this index according to the given probability and priority for constraintId. It returns the new index
// and whether it was possible at all to adjust this index.
func (i *index) adjust(constraint constraintId, probability Probability, priority Priority) (*index, bool) {
	if i.isBanned {
		return i, false
	}

	if probability == 0.0 {
		return bannedIndex, true
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
		probability:          i.probability,
		priority:             i.priority,
	}

	if shouldUpdateProbability {
		adjustedIndex.probabilityModifiers = make(map[constraintId]Probability)
		newProbability := float32(1.0)
		for k, prob := range i.probabilityModifiers {
			adjustedIndex.probabilityModifiers[k] = prob
			newProbability = newProbability * prob
		}
		adjustedIndex.probabilityModifiers[constraint] = probability
		adjustedIndex.probability = newProbability * probability
		adjustedIndex.isBanned = math.Abs(float64(adjustedIndex.probability)) < 1e-10
	}

	if shouldUpdatePriority {
		adjustedIndex.priorityModifiers = make(map[constraintId]Priority)
		newPriority := uint32(0)
		for k, prio := range i.priorityModifiers {
			adjustedIndex.priorityModifiers[k] = prio
			newPriority = newPriority + prio
		}
		adjustedIndex.priorityModifiers[constraint] = priority
		adjustedIndex.priority = newPriority + priority
	}

	return adjustedIndex, true
}

type packedProbPrio int64

func packPriorityProbability(probability float32, priority uint32) packedProbPrio {
	probabilityBits := math.Float32bits(probability)
	priorityBits := priority

	return packedProbPrio(uint64(probabilityBits)<<32 | uint64(priorityBits))
}

func unpackPriorityProbability(p packedProbPrio) (float32, uint32) {
	probabilityBits := uint32(p >> 32)
	priorityBits := uint32(p & 0xFFFFFFFF)

	return math.Float32frombits(probabilityBits), priorityBits
}
