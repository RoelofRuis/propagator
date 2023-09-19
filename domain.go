package propagator

import (
	"math"
)

// Domain represents a domain with states and their indices
type Domain struct {
	Name             string   // The name of the domain.
	indices          []*index // The indices in this domain.
	availableIndices []int
	minPriority      int     // The minimum priority over unbanned indices.
	sumProbability   float64 // The sum of probabilities over unbanned indices.
	entropy          float64 // The current entropy, or +Inf if not yet calculated.
	version          int     // Monotonically increasing version to track mutations.
}

// NewDomain initializes a new domain with a given name and probability distribution of its indices.
// The given distribution does not have to be normalized.
func NewDomain(name string, indices []*index) *Domain {
	domain := &Domain{
		Name:    name,
		indices: indices,
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
	if index >= len(d.indices) {
		return d.Contradict()
	}

	indices := make([]int, 0, len(d.availableIndices))
	for _, availableIndex := range d.availableIndices {
		if availableIndex == index {
			continue
		}
		indices = append(indices, availableIndex)
	}

	return d.Ban(indices...)
}

// Contradict returns the Mutation that bans all indices, forcing it to be in contradiction.
func (d *Domain) Contradict() Mutation {
	return d.Ban(d.availableIndices...)
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
	return len(d.availableIndices) > 1
}

// IsFixed returns whether the domain has only a single available state.
func (d *Domain) IsFixed() bool {
	return len(d.availableIndices) == 1
}

// IsContradiction returns whether the domain has no available states left to choose from.
func (d *Domain) IsContradiction() bool {
	return len(d.availableIndices) == 0
}

// WasUpdatedSince checks whether the domain was updated since the given version.
func (d *Domain) WasUpdatedSince(version int) bool {
	return d.version > version
}

// IndexPriority returns the priority of the given index.
func (d *Domain) IndexPriority(index int) int {
	return d.indices[index].priority
}

// IndexProbability returns the probability of the given index.
func (d *Domain) IndexProbability(index int) float64 {
	return d.indices[index].probability
}

// GetFixedIndex returns the fixed index for this domain or -1 if not yet fixed or contradictory.
func (d *Domain) GetFixedIndex() int {
	if !d.IsFixed() {
		return -1
	}
	return d.availableIndices[0]
}

// Entropy returns the entropy of this domain, taking into account the priorities of the indices.
// Only the indices with the non-banned highest priorities will be taken into account.
func (d *Domain) Entropy() float64 {
	// Calculate the entropy only once and then used cached version
	if !math.IsInf(d.entropy, +1) {
		return d.entropy
	}

	if d.sumProbability == 0.0 {
		d.entropy = math.Inf(-1)
		return d.entropy
	}

	entropy := 0.0
	for _, idx := range d.indices {
		if idx.isBanned || idx.priority != d.minPriority {
			continue
		}
		weightedProb := idx.probability / d.sumProbability
		entropy += weightedProb * math.Log2(weightedProb)
	}
	d.entropy = -entropy
	return d.entropy
}

// update recalculates internal domain state and mutates the version.
func (d *Domain) update() {
	d.version++

	availableIndexCount := 0
	d.sumProbability = 0.0
	d.minPriority = math.MaxInt
	d.entropy = math.Inf(+1)

	for _, idx := range d.indices {
		if idx.isBanned {
			continue
		}
		availableIndexCount++
		if idx.priority < d.minPriority {
			d.minPriority = idx.priority
		}
	}

	d.availableIndices = make([]int, 0, availableIndexCount)
	for i, idx := range d.indices {
		if !idx.isBanned {
			d.availableIndices = append(d.availableIndices, i)
		}
		if !idx.isBanned && idx.priority == d.minPriority {
			d.sumProbability += idx.probability
		}
	}
}
