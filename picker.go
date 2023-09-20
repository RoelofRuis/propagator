package propagator

import (
	"math"
	"math/rand"
)

// domainPicker selects the next domain for which a value will be picked.
type domainPicker func(m Model) Domain2

func nextDomainByMinEntropy(m Model) Domain2 {
	minEntropy := math.Inf(+1)
	var nextDomain Domain2
	for _, domain := range m.domains {
		if !domain.IsUnassigned() {
			continue
		}

		entropy := domain.Entropy()
		if entropy < minEntropy {
			nextDomain = domain
			minEntropy = entropy
		}
	}

	return nextDomain
}

func nextDomainByIndex(m Model) Domain2 {
	for _, domain := range m.domains {
		if domain.IsUnassigned() {
			return domain
		}
	}

	return nil
}

func nextDomainAtRandom(m Model) Domain2 {
	var validDomains []Domain2
	for _, domain := range m.domains {
		if domain.IsUnassigned() {
			validDomains = append(validDomains, domain)
		}
	}
	return validDomains[rand.Intn(len(validDomains))]
}

// indexPicker selects the next index from a given domain.
type indexPicker func(d Domain2) int

func nextIndexByProbability(d Domain2) int {
	cdfIdx := make([]int, 0, d.numIndices())
	cdf := make([]float64, 0, d.numIndices())

	minPriority := d.getMinPriority()

	probSum := 0.0
	prev := 0.0
	for i := 0; i < d.numIndices(); i++ {
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
