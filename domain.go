package propagator

import "math"

// Domain is a representation of a Variable domain.
// Use the mutator functions to create Mutation instances defining mutations on this domain.
type Domain struct {
	id    DomainId
	model *Model
}

// Assign creates a Mutation that assigns the given index to this domain.
func (d *Domain) Assign(index int) Mutation {
	if index >= len(d.model.domainIndices[d.id]) {
		return d.Contradict()
	}

	d.model.indexBuffer = d.model.indexBuffer[:0]
	for _, availableIndex := range d.model.domainAvailableIndices[d.id] {
		if availableIndex == index {
			continue
		}
		d.model.indexBuffer = append(d.model.indexBuffer, availableIndex)
	}

	return d.Exclude(d.model.indexBuffer...)
}

// Exclude creates a Mutation that excludes the given indices from this domain.
func (d *Domain) Exclude(indices ...int) Mutation {
	return d.Update(0.0, 0, indices...)
}

// Contradict creates a Mutation that excludes all indices from this domain.
func (d *Domain) Contradict() Mutation {
	d.model.indexBuffer = d.model.indexBuffer[:0]
	for _, availableIndex := range d.availableIndices() {
		d.model.indexBuffer = append(d.model.indexBuffer, availableIndex)
	}
	return d.Exclude(d.model.indexBuffer...)
}

// UpdatePriority creates a Mutation that updates the priority for the given indices.
func (d *Domain) UpdatePriority(value int, indices ...int) Mutation {
	return d.Update(1.0, value, indices...)
}

// UpdateProbability creates a Mutation that updates the probability for the given indices.
func (d *Domain) UpdateProbability(factor float64, indices ...int) Mutation {
	return d.Update(factor, 0, indices...)
}

// Update creates a Mutation that updates indices with the given probability and priority.
func (d *Domain) Update(probabilityFactor float64, priority int, indices ...int) Mutation {
	if len(indices) == 0 {
		return DoNothing
	}

	mutationIndices := make([]int, len(indices))
	copy(mutationIndices, indices)

	return Mutation{
		domain:      d,
		indices:     mutationIndices,
		probability: probabilityFactor,
		priority:    priority,
	}
}

// IsAssigned returns whether this domain is assigned exactly one index.
func (d *Domain) IsAssigned() bool {
	return len(d.availableIndices()) == 1
}

// IsUnassigned returns whether this domain allows a choice between more than one index.
func (d *Domain) IsUnassigned() bool {
	return len(d.availableIndices()) > 1
}

// IsInContradiction returns whether this domain has no indices available.
func (d *Domain) IsInContradiction() bool {
	return len(d.availableIndices()) == 0
}

// IndexPriority returns the priority of the given index.
func (d *Domain) IndexPriority(index int) int {
	return d.indices()[index].priority
}

// IndexProbability returns the probability of the given index.
func (d *Domain) IndexProbability(index int) float64 {
	return d.indices()[index].probability
}

// Name returns the name of this domain.
func (d *Domain) Name() string {
	return d.model.domainNames[d.id]
}

func (d *Domain) numIndices() int {
	return d.model.domainNumIndices[d.id]
}

func (d *Domain) version() int {
	return d.model.domainVersions[d.id]
}

func (d *Domain) getIndex(i int) *index {
	return d.model.domainIndices[d.id][i]
}

func (d *Domain) setIndex(i int, idx *index) {
	d.model.domainIndices[d.id][i] = idx
}

func (d *Domain) sumProbability() float64 {
	return d.model.domainSumProbability[d.id]
}

func (d *Domain) minPriority() int {
	return d.model.domainMinPriority[d.id]
}

func (d *Domain) indices() []*index {
	return d.model.domainIndices[d.id]
}

func (d *Domain) availableIndices() []int {
	return d.model.domainAvailableIndices[d.id]
}

func (d *Domain) entropy() float64 {
	if !math.IsInf(d.model.domainEntropy[d.id], +1) {
		return d.model.domainEntropy[d.id]
	}

	if d.sumProbability() == 0.0 {
		d.model.domainEntropy[d.id] = math.Inf(-1)
		return d.model.domainEntropy[d.id]
	}

	entropy := 0.0
	for _, idx := range d.indices() {
		if idx.isBanned || idx.priority != d.minPriority() {
			continue
		}
		weightedProb := idx.probability / d.sumProbability()
		entropy += weightedProb * math.Log2(weightedProb)
	}
	d.model.domainEntropy[d.id] = -entropy
	return d.model.domainEntropy[d.id]
}

// update is called internally after applying a mutation.
// It resets internal domain state and precalculate values.
func (d *Domain) update() {
	d.model.domainVersions[d.id]++
	d.model.domainSumProbability[d.id] = 0.0
	d.model.domainMinPriority[d.id] = math.MaxInt
	d.model.domainEntropy[d.id] = math.Inf(+1)
	d.model.domainAvailableIndices[d.id] = d.model.domainAvailableIndices[d.id][:0]

	for i, idx := range d.model.domainIndices[d.id] {
		if !idx.isBanned {
			d.model.domainAvailableIndices[d.id] = append(d.model.domainAvailableIndices[d.id], i)
		}
		if !idx.isBanned && idx.priority < d.model.domainMinPriority[d.id] {
			d.model.domainMinPriority[d.id] = idx.priority
		}
	}

	for _, idx := range d.model.domainIndices[d.id] {
		if !idx.isBanned && idx.priority == d.model.domainMinPriority[d.id] {
			d.model.domainSumProbability[d.id] += idx.probability
		}
	}
}
