package main

import (
	"github.com/RoelofRuis/propagator"
	"math"
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

type FixedCage struct {
	Cell  *Cell
	Value int
}

func (c FixedCage) GetLinkedDomains() []*propagator.Domain {
	return []*propagator.Domain{c.Cell.Domain}
}

func (c FixedCage) Propagate(mutator *propagator.Mutator) {
	mutator.Add(c.Cell.FixByValue(c.Value))
}

type SumCage struct {
	Cells []*Cell
	Value int
}

func (c SumCage) GetLinkedDomains() []*propagator.Domain {
	return propagator.DomainsOf(c.Cells)
}

func (c SumCage) Propagate(mutator *propagator.Mutator) {
	total := 0
	numFixed := 0
	for _, cell := range c.Cells {
		if !cell.IsFixed() {
			continue
		}
		numFixed += 1
		total += cell.GetFixedValue()
	}
	if numFixed == (len(c.Cells) - 1) {
		for _, cell := range c.Cells {
			if cell.IsFixed() {
				continue
			}
			mutator.Add(cell.FixByValue(c.Value - total))
		}
	} else if numFixed == len(c.Cells) && total != c.Value {
		mutator.Add(c.Cells[0].Contradict())
	}
}

type ProdCage struct {
	Cells []*Cell
	Value int
}

func (c ProdCage) GetLinkedDomains() []*propagator.Domain {
	return propagator.DomainsOf(c.Cells)
}

func (c ProdCage) Propagate(mutator *propagator.Mutator) {
	total := 1
	for _, cell := range c.Cells {
		if !cell.IsFixed() {
			mutator.Add(cell.BanBy(func(i int) bool {
				return c.Value%i != 0
			}))
			return
		}
		total *= cell.GetFixedValue()
	}
	if total != c.Value {
		mutator.Add(c.Cells[0].Contradict())
	}
}

type SubCage struct {
	Cells []*Cell
	Value int
}

func (c SubCage) GetLinkedDomains() []*propagator.Domain {
	if len(c.Cells) > 3 {
		// FIXME: this assumes only 2 cells take part
		panic("subtractive cages with more than two cells not supported")
	}
	return propagator.DomainsOf(c.Cells)
}

func (c SubCage) Propagate(mutator *propagator.Mutator) {
	var values []int
	for _, c := range c.Cells {
		if !c.IsFixed() {
			continue
		}
		values = append(values, c.GetFixedValue())
	}

	if len(c.Cells) == 2 && len(values) == 1 {
		for _, cell := range c.Cells {
			if cell.IsFixed() {
				continue
			}
			sum := c.Value + values[0]
			diff := values[0] - c.Value
			mutator.Add(cell.BanBy(func(i int) bool {
				return i != sum && i != diff
			}))
		}
	} else if len(c.Cells) == 2 && len(values) == 2 {
		firstMiss := values[0]-values[1] != c.Value
		secondMiss := values[1]-values[0] != c.Value
		if firstMiss && secondMiss {
			mutator.Add(c.Cells[0].Contradict())
		}
	} else if len(c.Cells) == 3 && len(values) == 3 {
		miss1 := values[0]-values[1]-values[2] != c.Value
		miss2 := values[0]-values[2]-values[1] != c.Value
		miss3 := values[1]-values[0]-values[2] != c.Value
		miss4 := values[1]-values[2]-values[0] != c.Value
		miss5 := values[2]-values[0]-values[1] != c.Value
		miss6 := values[2]-values[1]-values[0] != c.Value
		if miss1 && miss2 && miss3 && miss4 && miss5 && miss6 {
			mutator.Add(c.Cells[0].Contradict())
		}
	}
}

type DivCage struct {
	Cells []*Cell
	Value int
}

func (c DivCage) GetLinkedDomains() []*propagator.Domain {
	if len(c.Cells) > 2 {
		// FIXME: this assumes only 2 cells take part
		panic("division cages with more than two cells not supported")
	}
	return propagator.DomainsOf(c.Cells)
}

func (c DivCage) Propagate(mutator *propagator.Mutator) {
	var values []int
	for _, c := range c.Cells {
		if !c.IsFixed() {
			return
		}
		values = append(values, c.GetFixedValue())
	}
	firstHit := math.Abs((float64(values[0])/float64(values[1]))-float64(c.Value)) < 10e-10
	secondHit := math.Abs((float64(values[1])/float64(values[0]))-float64(c.Value)) < 10e-10
	if !firstHit && !secondHit {
		mutator.Add(c.Cells[0].Contradict())
	}
}
