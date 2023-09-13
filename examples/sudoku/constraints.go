package main

import (
	"github.com/RoelofRuis/propagator"
)

type House struct {
	Cells []*Cell
}

func (h House) GetLinkedDomains() []*propagator.Domain {
	return propagator.DomainsOf(h.Cells)
}

func (h House) Propagate(mutator *propagator.Mutator) {
	for _, i := range h.Cells {
		if !i.IsFixed() {
			continue
		}

		for _, j := range h.Cells {
			if j == i {
				continue
			}
			mutator.Add(j.BanByValue(i.GetFixedValue()))
		}
	}
}
