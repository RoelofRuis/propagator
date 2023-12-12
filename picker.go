package propagator

import (
	"math"
	"math/rand"
)

// domainPicker selects the next domain for which a value will be picked.
type domainPicker func(m Model) *Domain

func nextDomainByMinEntropy(m Model) *Domain {
	minEntropy := math.Inf(+1)
	var nextDomain *Domain
	for _, domain := range m.domains {
		if !domain.IsUnassigned() {
			continue
		}

		entropy := m.EntropyOf(domain)
		if entropy < minEntropy {
			nextDomain = domain
			minEntropy = entropy
		}
	}

	return nextDomain
}

func nextDomainByIndex(m Model) *Domain {
	for _, domain := range m.domains {
		if domain.IsUnassigned() {
			return domain
		}
	}

	return nil
}

func nextDomainAtRandom(m Model) *Domain {
	var validDomains []*Domain
	for _, domain := range m.domains {
		if domain.IsUnassigned() {
			validDomains = append(validDomains, domain)
		}
	}
	return validDomains[rand.Intn(len(validDomains))]
}

// indexPicker selects the next index from a given domain.
type indexPicker func(m Model, d *Domain) int

func nextIndexByProbability(m Model, d *Domain) int {
	cdfIdx := make([]int, 0, m.NumIndicesOf(d))
	cdf := make([]float64, 0, m.NumIndicesOf(d))

	minPriority := d.minPriority

	probSum := 0.0
	prev := 0.0
	for i := 0; i < m.NumIndicesOf(d); i++ {
		idx := d.getIndex(i)
		if idx.isBanned || idx.priority != minPriority {
			continue
		}

		cdfIdx = append(cdfIdx, i)
		nextProb := prev + idx.probability
		cdf = append(cdf, nextProb)
		prev = nextProb
		probSum += idx.probability
	}

	if len(cdf) == 0 {
		return -1
	}

	r := rand.Float64() * probSum
	idx := 0
	for r > cdf[idx] {
		idx++
	}

	return cdfIdx[idx]
}
