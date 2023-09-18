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

	probStr := strconv.FormatFloat(probability, 'd', -1, 64)
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
