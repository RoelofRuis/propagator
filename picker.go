package propagator

import (
	"math"
	"math/rand"
	"slices"
)

// domainPicker selects the next domain for which a value will be picked.
type domainPicker func(m Model, rnd *rand.Rand) *Domain

// nextDomainByMinEntropy selects the next domain that has minimal Shannon entropy.
func nextDomainByMinEntropy(m Model, rnd *rand.Rand) *Domain {
	minEntropy := math.Inf(+1)
	var nextDomain *Domain
	for _, domain := range m.Domains {
		if !domain.canBePicked() {
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

// nextDomainByIndex selects the next unassigned domain from the list in the order as they were inserted into the model.
func nextDomainByIndex(m Model, rnd *rand.Rand) *Domain {
	for _, domain := range m.Domains {
		if domain.canBePicked() {
			return domain
		}
	}

	return nil
}

// nextDomainAtRandom selects the next domain at random.
func nextDomainAtRandom(m Model, rnd *rand.Rand) *Domain {
	var validDomains []*Domain
	for _, domain := range m.Domains {
		if domain.canBePicked() {
			validDomains = append(validDomains, domain)
		}
	}
	return validDomains[rnd.Intn(len(validDomains))]
}

// indexPicker selects the next index from a given domain.
type indexPicker interface {
	init(m Model, rnd *rand.Rand)
	nextIndex(d *Domain) int
}

type ProbabilisticIndexPicker struct {
	// cumulative distribution function index
	cdfIdx []int
	// cumulative distribution function
	cdf []float64
	// pointer to the random number generator
	rnd *rand.Rand
}

func (p *ProbabilisticIndexPicker) init(m Model, rnd *rand.Rand) {
	maxIndices := slices.Max(m.domainNumIndices)
	p.cdfIdx = make([]int, 0, maxIndices)
	p.cdf = make([]float64, 0, maxIndices)
	p.rnd = rnd
}

func (p *ProbabilisticIndexPicker) nextIndex(d *Domain) int {
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

	r := p.rnd.Float64() * probSum
	idx := 0
	for r > p.cdf[idx] {
		idx++
	}

	return p.cdfIdx[idx]
}
