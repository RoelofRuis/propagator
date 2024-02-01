package propagator

import (
	"math"
	"math/rand"
	"slices"
)

// domainPicker selects the next domain for which a value will be picked.
type domainPicker interface {
	init(m Model, rnd *rand.Rand)
	nextDomain(m Model) *Domain
}

// MinRemainingValuesPicker selects from the unassigned domains the domain that has the fewest legal values.
// As a tie-breaker among the Minimum Remaining Values variables it picks the variable with the most constraints on
// remaining variables.
type MinRemainingValuesPicker struct {
	candidates []*Domain
}

func (p *MinRemainingValuesPicker) init(m Model, rnd *rand.Rand) {
	p.candidates = make([]*Domain, 0, len(m.Domains))
}

func (p *MinRemainingValuesPicker) nextDomain(m Model) *Domain {
	p.candidates = p.candidates[:0]
	minIndices := math.MaxInt
	for _, domain := range m.Domains {
		if !domain.CanBePicked() {
			continue
		}
		numIndices := len(domain.AvailableIndices())
		if numIndices < minIndices {
			minIndices = numIndices
			p.candidates = p.candidates[:0]
		}
		if numIndices == minIndices {
			p.candidates = append(p.candidates, domain)
		}
	}

	if len(p.candidates) == 0 {
		return nil
	}
	if len(p.candidates) == 1 {
		return p.candidates[0]
	}

	maxConstraints := 0
	var nextDomain *Domain
	for _, candidate := range p.candidates {
		constraintCount := candidate.numRelevantConstraints()
		if constraintCount > maxConstraints {
			maxConstraints = constraintCount
			nextDomain = candidate
		}
	}

	return nextDomain
}

// MinEntropyDomainPicker selects from the unassigned domains the domain that has minimal Shannon entropy.
type MinEntropyDomainPicker struct{}

func (p *MinEntropyDomainPicker) init(m Model, rnd *rand.Rand) {}

func (p *MinEntropyDomainPicker) nextDomain(m Model) *Domain {
	minEntropy := math.Inf(+1)
	var nextDomain *Domain
	for _, domain := range m.Domains {
		if !domain.CanBePicked() {
			continue
		}

		entropy := domain.entropy()
		if entropy < minEntropy {
			nextDomain = domain
			minEntropy = entropy
		}
	}

	return nextDomain
}

// IndexDomainPicker selects the next unassigned domain in the order they were inserted into the model.
type IndexDomainPicker struct{}

func (p *IndexDomainPicker) init(m Model, rnd *rand.Rand) {}

func (p *IndexDomainPicker) nextDomain(m Model) *Domain {
	for _, domain := range m.Domains {
		if domain.CanBePicked() {
			return domain
		}
	}

	return nil
}

// RandomDomainPicker selects the next unassigned domain at random.
type RandomDomainPicker struct {
	rnd *rand.Rand
}

func (p *RandomDomainPicker) init(m Model, rnd *rand.Rand) {
	p.rnd = rnd
}

func (p *RandomDomainPicker) nextDomain(m Model) *Domain {
	var validDomains []*Domain
	for _, domain := range m.Domains {
		if domain.CanBePicked() {
			validDomains = append(validDomains, domain)
		}
	}
	return validDomains[p.rnd.Intn(len(validDomains))]
}

// indexPicker selects the next index from a given domain.
type indexPicker interface {
	init(m Model, rnd *rand.Rand)
	nextIndex(d *Domain) int
}

// LeastConstrainingValueIndexPicker selects the value that rules out the fewest values in the remaining variables.
type LeastConstrainingValueIndexPicker struct {
}

func (p *LeastConstrainingValueIndexPicker) init(m Model, rnd *rand.Rand) {}

func (p *LeastConstrainingValueIndexPicker) nextIndex(d *Domain) int {
	panic("not implemented") // TODO: implement
}

// RandomIndexPicker selects the next index at random without using probability and priority information.
type RandomIndexPicker struct {
	rnd *rand.Rand
}

func (p *RandomIndexPicker) init(m Model, rnd *rand.Rand) {
	p.rnd = rnd
}

func (p *RandomIndexPicker) nextIndex(d *Domain) int {
	indices := d.AvailableIndices()
	if len(indices) == 0 {
		return -1
	}
	return indices[p.rnd.Intn(len(indices))]
}

// ProbabilisticIndexPicker selects the next index based on chance, taking into account the probabilities of the
// individual values. It also incorporates priority, picking only from the group of values with the smallest priority
// value.
type ProbabilisticIndexPicker struct {
	// cumulative distribution function index
	cdfIdx []int
	// cumulative distribution function
	cdf []Probability

	rnd *rand.Rand
}

func (p *ProbabilisticIndexPicker) init(m Model, rnd *rand.Rand) {
	maxIndices := slices.Max(m.domainNumIndices)
	p.cdfIdx = make([]int, 0, maxIndices)
	p.cdf = make([]Probability, 0, maxIndices)
	p.rnd = rnd
}

func (p *ProbabilisticIndexPicker) nextIndex(d *Domain) int {
	p.cdfIdx = p.cdfIdx[:0]
	p.cdf = p.cdf[:0]

	minPriority := d.minPriority()

	probSum := float32(0.0)
	prev := float32(0.0)
	for i := 0; i < d.numIndices(); i++ {
		idx := d.getIndex(i)
		if idx == nil || idx.priority != minPriority {
			continue
		}

		p.cdfIdx = append(p.cdfIdx, i)
		nextProb := prev + idx.probability
		p.cdf = append(p.cdf, nextProb)
		prev = nextProb
		probSum += idx.probability
	}

	if len(p.cdf) == 0 {
		return -1
	}

	r := p.rnd.Float32() * probSum
	idx := 0
	for r > p.cdf[idx] {
		idx++
	}

	return p.cdfIdx[idx]
}
