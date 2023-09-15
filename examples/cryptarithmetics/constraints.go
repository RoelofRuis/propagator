package main

import "github.com/RoelofRuis/propagator"

type Sum struct {
	A   Number
	B   Number
	Sum Number
}

func (s Sum) GetLinkedDomains() []*propagator.Domain {
	aDomains := s.A.GetLinkedDomains()
	bDomains := s.B.GetLinkedDomains()
	sumDomains := s.Sum.GetLinkedDomains()
	res := make([]*propagator.Domain, len(aDomains)+len(bDomains)+len(sumDomains))
	copy(res[0:], aDomains)
	copy(res[len(aDomains):], bDomains)
	copy(res[len(aDomains)+len(bDomains):], sumDomains)
	return res
}

func (s Sum) Propagate(m *propagator.Mutator) {
	if s.A.IsFixed() && s.B.IsFixed() && s.Sum.IsFixed() {
		if s.A.Decimal()+s.B.Decimal() != s.Sum.Decimal() {
			s.A.Contradict(m)
			s.B.Contradict(m)
			s.Sum.Contradict(m)
		}
	}
}

type Number struct {
	Variables []*propagator.Variable[int]
}

func (n Number) GetLinkedDomains() []*propagator.Domain {
	return propagator.DomainsOf(n.Variables)
}

func (n Number) Propagate(m *propagator.Mutator) {
	if !n.Variables[0].IsFixed() {
		m.Add(n.Variables[0].BanByValue(0))
	}
}

func (n Number) IsFixed() bool {
	for _, v := range n.Variables {
		if !v.IsFixed() {
			return false
		}
	}
	return true
}

func (n Number) Contradict(m *propagator.Mutator) {
	for _, v := range n.Variables {
		m.Add(v.Contradict())
	}
}

func (n Number) Decimal() int {
	sum := 0
	numDigits := len(n.Variables)
	for i := 0; i < numDigits; i++ {
		pow := 1
		for n := i + 1; n < numDigits; n++ {
			pow *= 10
		}
		sum += n.Variables[i].GetFixedValue() * pow
	}
	return sum
}

type AllDifferent struct {
	Variables []*propagator.Variable[int]
}

func (a AllDifferent) GetLinkedDomains() []*propagator.Domain {
	return propagator.DomainsOf(a.Variables)
}

func (a AllDifferent) Propagate(m *propagator.Mutator) {
	for _, v := range a.Variables {
		if v.IsFixed() {
			for _, w := range a.Variables {
				if w != v {
					m.Add(w.BanByValue(v.GetFixedValue()))
				}
			}
		}
	}
}
