package main

import (
	"github.com/RoelofRuis/propagator"
)

type House struct {
	Cells []*Cell
}

func (h House) Scope() []propagator.Domain2 {
	var l []propagator.Domain2
	for _, c := range h.Cells {
		l = append(l, propagator.Domain2(c))
	}
	return l
}

func (h House) Propagate(mutator *propagator.Mutator) {
	for _, i := range h.Cells {
		if !i.IsAssigned() {
			continue
		}

		for _, j := range h.Cells {
			if j == i {
				continue
			}
			mutator.Add(j.ExcludeByValue(i.GetAssignedValue()))
		}
	}
}
