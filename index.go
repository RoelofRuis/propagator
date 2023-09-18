package propagator

import (
	"math"
	"strconv"
)

var (
	indexFactorySingleton = &indexFactory{
		indices: make(map[string]*index),
	}
	bannedIndex = &index{
		probabilityModifiers: nil,
		priorityModifiers:    nil,
		probability:          0.0,
		priority:             math.MaxInt,
		isBanned:             true,
	}
)

type indexFactory struct {
	indices map[string]*index
}

func (f *indexFactory) create(probability float64, priority int) *index {
	if math.Abs(probability) < 1e-10 {
		return bannedIndex
	}

	probStr := strconv.FormatFloat(probability, 'f', -1, 64)
	prioStr := strconv.FormatInt(int64(priority), 10)
	hash := probStr + prioStr

	storedIndex, has := f.indices[hash]
	if !has {
		storedIndex = &index{
			probabilityModifiers: map[constraintId]float64{-1: probability},
			priorityModifiers:    map[constraintId]int{-1: priority},
			probability:          probability,
			priority:             priority,
			isBanned:             false,
		}
		f.indices[hash] = storedIndex
	}
	return storedIndex
}

// index stores the probability and priority of a single index in a domain.
// The -1 key in both maps serves as the respective base probability and base priority. These are fixed and cannot be
// modified by a constraint.
type index struct {
	probabilityModifiers map[constraintId]float64
	priorityModifiers    map[constraintId]int

	// Product of probability modifiers
	probability float64
	// Sum of priority modifiers
	priority int
	// Whether the index is currently banned
	isBanned bool
}

func (i *index) adjust(constraint constraintId, probability float64, priority int) (*index, bool) {
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
		adjustedIndex.probabilityModifiers = make(map[constraintId]float64)
		newProbability := 1.0
		for k, prob := range i.probabilityModifiers {
			adjustedIndex.probabilityModifiers[k] = prob
			newProbability = newProbability * prob
		}
		adjustedIndex.probabilityModifiers[constraint] = probability
		adjustedIndex.probability = newProbability * probability
		adjustedIndex.isBanned = math.Abs(adjustedIndex.probability) < 1e-10
	}

	if shouldUpdatePriority {
		adjustedIndex.priorityModifiers = make(map[constraintId]int)
		newPriority := 0
		for k, prio := range i.priorityModifiers {
			adjustedIndex.priorityModifiers[k] = prio
			newPriority = newPriority + prio
		}
		adjustedIndex.priorityModifiers[constraint] = priority
		adjustedIndex.priority = newPriority + priority
	}

	return adjustedIndex, true
}
