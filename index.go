package propagator

import (
	"math"
)

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

var (
	bannedIndex = &index{
		probabilityModifiers: nil,
		priorityModifiers:    nil,
		probability:          0.0,
		priority:             math.MaxInt,
		isBanned:             true,
	}
)

// FIXME: this is memory consuming
func newIndex(baseProbability float64, priority int) *index {
	return &index{
		probabilityModifiers: map[constraintId]float64{-1: baseProbability},
		priorityModifiers:    map[constraintId]int{-1: priority},
		probability:          baseProbability,
		priority:             priority,
		isBanned:             math.Abs(baseProbability) < 1e-10,
	}
}

// FIXME: this is memory consuming
func (i *index) adjust(constraint constraintId, probability float64, priority int) (*index, bool) {
	if i.isBanned {
		return i, false
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

	if probability == 0.0 {
		return bannedIndex, true
	}

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
