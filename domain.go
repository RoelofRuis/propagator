package propagator

import (
	"math"
)

type DomainState int

const (
	Free          DomainState = 0
	Fixed         DomainState = 1
	Contradiction DomainState = 2
)

// Domain represents a domain with states and their indices
type Domain struct {
	Name    string      // The name of the domain.
	indices []index     // The state of the available indices in this domain.
	state   DomainState // The current state the domain is in.
	version int         // Monotonically increasing version to track mutations.
}

// NewDomain initializes a new domain with a given name and probability distribution of its indices.
// The given distribution does not have to be normalized.
func NewDomain(name string, indices []index) *Domain {
	domain := &Domain{
		Name:    name,
		indices: indices,
		state:   Free,
		version: 0,
	}

	domain.update()

	return domain
}

// Ban returns the Mutation that bans the given indices.
func (d *Domain) Ban(indices ...int) Mutation {
	return d.Update(0.0, 0, indices...)
}

// Fix returns the Mutation that fixes this domain to the given index.
func (d *Domain) Fix(index int) Mutation {
	var indices []int
	for i := range d.indices {
		if i != index {
			indices = append(indices, i)
		}
	}
	return d.Ban(indices...)
}

// Contradict returns the Mutation that bans all indices, forcing it to be in contradiction.
func (d *Domain) Contradict() Mutation {
	indices := make([]int, len(d.indices))
	for i := range d.indices {
		indices[i] = i
	}
	return d.Ban(indices...)
}

// UpdatePriority returns the Mutation that changes the priority of the given indices.
func (d *Domain) UpdatePriority(value int, indices ...int) Mutation {
	return d.Update(1.0, value, indices...)
}

// UpdateProbability returns the Mutation that adjusts the probability of the given indices by multiplying
// them with the given factor.
func (d *Domain) UpdateProbability(factor float64, indices ...int) Mutation {
	return d.Update(factor, 0, indices...)
}

// Update returns the Mutation that changes the given probability and priority for the indicated indices.
func (d *Domain) Update(probabilityFactor float64, priority int, indices ...int) Mutation {
	if len(indices) == 0 {
		return DoNothing
	}
	return Mutation{
		domain:      d,
		indices:     indices,
		probability: probabilityFactor,
		priority:    priority,
	}
}

// IsFree returns whether the domain has more than one single available state.
func (d *Domain) IsFree() bool {
	return d.state == Free
}

// IsFixed returns whether the domain has only a single available state.
func (d *Domain) IsFixed() bool {
	return d.state == Fixed
}

// IsContradiction returns whether the domain has no available states left to choose from.
func (d *Domain) IsContradiction() bool {
	return d.state == Contradiction
}

// WasUpdatedSince checks whether the domain was updated since the given version.
func (d *Domain) WasUpdatedSince(version int) bool {
	return d.version > version
}

// IndexIsBanned returns whether the given index state is banned.
func (d *Domain) IndexIsBanned(index int) bool {
	return d.indices[index].isBanned()
}

// IndexPriority returns the priority of the given index.
func (d *Domain) IndexPriority(index int) int {
	return d.indices[index].priority()
}

// IndexProbability returns the probability of the given index.
func (d *Domain) IndexProbability(index int) float64 {
	return d.indices[index].probability()
}

// GetFixedIndex returns the fixed index for this domain or -1 if not yet fixed or contradictory.
func (d *Domain) GetFixedIndex() int {
	if !d.IsFixed() {
		return -1
	}
	for i := range d.indices {
		if !d.IndexIsBanned(i) {
			return i
		}
	}
	return -1
}

// Entropy returns the entropy of this domain, taking into account the priorities of the indices.
// Only the indices with the non-banned highest priorities will be taken into account.
func (d *Domain) Entropy() float64 {
	minPriority := d.MinPriority()

	probSum := 0.0
	for _, idx := range d.indices {
		if idx.isBanned() || idx.priority() != minPriority {
			continue
		}
		probSum += idx.probability()
	}

	if probSum == 0.0 {
		return math.Inf(-1)
	}

	entropy := 0.0
	for _, idx := range d.indices {
		if idx.isBanned() || idx.priority() != minPriority {
			continue
		}
		weightedProb := idx.probability() / probSum
		entropy += weightedProb * math.Log2(weightedProb)
	}
	return -entropy
}

// MinPriority returns the minimal priority value that has unbanned indices. It will return math.MaxInt if no priority is available.
func (d *Domain) MinPriority() int {
	minPriority := math.MaxInt
	for _, idx := range d.indices {
		if idx.isBanned() {
			continue
		}
		idxPrio := idx.priority()
		if idxPrio < minPriority {
			minPriority = idxPrio
		}
	}
	return minPriority
}

// update recalculates internal domain state and mutations the version.
func (d *Domain) update() {
	d.version++

	d.state = Contradiction
	for i := range d.indices {
		if !d.IndexIsBanned(i) {
			if d.state == Contradiction {
				d.state = Fixed
			} else {
				d.state = Free
				break
			}
		}
	}
}
