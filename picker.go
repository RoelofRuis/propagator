package propagator

import (
	"math"
	"math/rand"
)

// domainPicker selects the next domain to be collapsed.
type domainPicker func(m Model) *Domain

func nextDomainByMinEntropy(m Model) *Domain {
	minEntropy := math.Inf(+1)
	var nextDomain *Domain
	for _, domain := range m.domains {
		if !domain.IsFree() {
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

func nextDomainByIndex(m Model) *Domain {
	for _, domain := range m.domains {
		if domain.IsFree() {
			return domain
		}
	}

	return nil
}

// indexPicker selects the next index from a given domain.
type indexPicker func(d *Domain) int

func nextIndexByProbability(d *Domain) int {
	cdfIdx := make([]int, 0, len(d.indices))
	cdf := make([]float64, 0, len(d.indices))

	minPriority := d.MinPriority()

	probSum := 0.0
	prev := 0.0
	for i, idx := range d.indices {
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
