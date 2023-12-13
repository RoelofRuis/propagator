package propagator

import (
	"math"
	"math/rand"
	"slices"
)

// domainPicker selects the next domain for which a value will be picked.
type domainPicker func(m Model, rnd *rand.Rand) *Domain

func nextDomainByMinEntropy(m Model, rnd *rand.Rand) *Domain {
	minEntropy := math.Inf(+1)
	var nextDomain *Domain
	for _, domain := range m.Domains {
		if !domain.IsUnassigned() {
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

func nextDomainByIndex(m Model, rnd *rand.Rand) *Domain {
	for _, domain := range m.Domains {
		if domain.IsUnassigned() {
			return domain
		}
	}

	return nil
}

func nextDomainAtRandom(m Model, rnd *rand.Rand) *Domain {
	var validDomains []*Domain
	for _, domain := range m.Domains {
		if domain.IsUnassigned() {
			validDomains = append(validDomains, domain)
		}
	}
	return validDomains[rnd.Intn(len(validDomains))]
}

// indexPicker selects the next index from a given domain.
type indexPicker interface {
	init(m Model)
	nextIndex(d *Domain, rnd *rand.Rand) int
}

type ProbabilisticIndexPicker struct {
	// cumulative distribution function index
	cdfIdx []int
	// cumulative distribution function
	cdf []float64
}

func (p *ProbabilisticIndexPicker) init(m Model) {
	maxIndices := slices.Max(m.domainNumIndices)
	p.cdfIdx = make([]int, 0, maxIndices)
	p.cdf = make([]float64, 0, maxIndices)
}

func (p *ProbabilisticIndexPicker) nextIndex(d *Domain, rnd *rand.Rand) int {
	p.cdfIdx = p.cdfIdx[:0]
	p.cdf = p.cdf[:0]

	minPriority := d.minPriority()

	probSum := 0.0
	prev := 0.0
	for i := 0; i < d.numIndices(); i++ {
		idx := d.getIndex(i)
		if idx.isBanned || idx.priority != minPriority {
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

	r := rnd.Float64() * probSum
	idx := 0
	for r > p.cdf[idx] {
		idx++
	}

	return p.cdfIdx[idx]
}
