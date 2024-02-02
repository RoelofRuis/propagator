package propagator

import (
	"math"
)

// Domain is a representation of a Variable domain.
// Use the mutator functions to create Mutation instances defining mutations on this domain.
type Domain struct {
	id    DomainId
	model *Model
}

// Assign creates a Mutation that assigns the given index to this domain.
func (d *Domain) Assign(index int) Mutation {
	if index >= d.model.domainNumIndices[d.id] {
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
	for _, availableIndex := range d.AvailableIndices() {
		d.model.indexBuffer = append(d.model.indexBuffer, availableIndex)
	}
	return d.Exclude(d.model.indexBuffer...)
}

// UpdatePriority creates a Mutation that updates the priority for the given indices.
func (d *Domain) UpdatePriority(value Priority, indices ...int) Mutation {
	return d.Update(1.0, value, indices...)
}

// UpdateProbability creates a Mutation that updates the probability for the given indices.
func (d *Domain) UpdateProbability(factor Probability, indices ...int) Mutation {
	return d.Update(factor, 0, indices...)
}

// Update creates a Mutation that updates indices with the given probability and priority.
func (d *Domain) Update(factor Probability, priority Priority, indices ...int) Mutation {
	if len(indices) == 0 {
		return DoNothing
	}

	mutationIndices := make([]int, len(indices))
	copy(mutationIndices, indices)

	return Mutation{
		domain:      d,
		indices:     mutationIndices,
		probability: factor,
		priority:    priority,
	}
}

// IsAssigned returns whether this domain is assigned exactly one index.
func (d *Domain) IsAssigned() bool {
	return len(d.AvailableIndices()) == 1
}

// IsUnassigned returns whether this domain allows a choice between more than one index.
func (d *Domain) IsUnassigned() bool {
	return len(d.AvailableIndices()) > 1
}

// IsInContradiction returns whether this domain has no indices available.
func (d *Domain) IsInContradiction() bool {
	return len(d.AvailableIndices()) == 0
}

// IndexPriority returns the priority of the given index.
func (d *Domain) IndexPriority(index int) Priority {
	return d.model.domainIndexPriority[d.id][index]
}

// IndexProbability returns the probability of the given index.
func (d *Domain) IndexProbability(index int) Probability {
	return d.model.domainIndexProbability[d.id][index]
}

// Name returns the name of this domain.
func (d *Domain) Name() string {
	return d.model.domainNames[d.id]
}

// AvailableIndices returns the indices that can still be selected for this domain.
func (d *Domain) AvailableIndices() []int {
	return d.model.domainAvailableIndices[d.id]
}

// IsHidden returns whether this domain is hidden and will not be picked and solved for.
// Its constraint values will still be propagated though; in this way it still is part of the problem space.
func (d *Domain) IsHidden() bool {
	return d.model.domainHidden[d.id]
}

// CanBePicked returns whether this domain may be selected to be assigned a value.
// The main use is in the picking algorithms to check available domains.
func (d *Domain) CanBePicked() bool {
	return d.IsUnassigned() && !d.IsHidden()
}

func (d *Domain) version() int {
	return d.model.domainVersions[d.id]
}

func (d *Domain) sumProbability() Probability {
	return d.model.domainSumProbability[d.id]
}

func (d *Domain) minPriority() Priority {
	return d.model.domainMinPriority[d.id]
}

// numRelevantConstraints returns the number of constraints that this domain shares with other domains that are still
// unassigned.
func (d *Domain) numRelevantConstraints() int {
	count := 0
iterateConstraints:
	for _, constraintId := range d.model.domainConstraints[d.id] {
		constraint := d.model.constraints[constraintId]
		for _, linkedId := range constraint.linkedDomains {
			if linkedId == d.id {
				continue
			}
			if len(d.model.domainAvailableIndices[linkedId]) > 1 {
				count++
				continue iterateConstraints
			}
		}
	}
	return count
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
	for i := 0; i < d.model.domainNumIndices[d.id]; i++ {
		idxProbability := d.model.domainIndexProbability[d.id][i]

		if idxProbability < 10e-10 {
			continue
		}

		idxPriority := d.model.domainIndexPriority[d.id][i]

		if idxPriority != d.minPriority() {
			continue
		}

		weightedProb := idxProbability / d.sumProbability()
		entropy += float64(weightedProb) * math.Log2(float64(weightedProb))
	}
	d.model.domainEntropy[d.id] = -entropy
	return d.model.domainEntropy[d.id]
}

// update is called internally after applying a mutation.
// It resets internal domain state and precalculate values.
func (d *Domain) update() {
	d.model.domainVersions[d.id]++
	d.model.domainSumProbability[d.id] = 0.0
	d.model.domainMinPriority[d.id] = math.MaxUint32
	d.model.domainEntropy[d.id] = math.Inf(+1)
	d.model.domainAvailableIndices[d.id] = d.model.domainAvailableIndices[d.id][:0]

	for i := 0; i < d.model.domainNumIndices[d.id]; i++ {
		d.model.domainIndexProbability[d.id][i] = d.model.domainIndexProbabilityModifiers[d.id][i].Product()
		d.model.domainIndexPriority[d.id][i] = d.model.domainIndexPriorityModifiers[d.id][i].Sum()
	}

	for i := 0; i < d.model.domainNumIndices[d.id]; i++ {
		idxProbability := d.model.domainIndexProbability[d.id][i]
		if idxProbability < 10e-10 {
			continue
		}

		d.model.domainAvailableIndices[d.id] = append(d.model.domainAvailableIndices[d.id], i)

		idxPriority := d.model.domainIndexPriority[d.id][i]

		if idxPriority < d.model.domainMinPriority[d.id] {
			d.model.domainMinPriority[d.id] = idxPriority
		}
	}

	for i := 0; i < d.model.domainNumIndices[d.id]; i++ {
		idxProbability := d.model.domainIndexProbability[d.id][i]
		if idxProbability < 10e-10 {
			continue
		}

		idxPriority := d.model.domainIndexPriority[d.id][i]

		if idxPriority == d.model.domainMinPriority[d.id] {
			d.model.domainSumProbability[d.id] += idxProbability
		}
	}
}
