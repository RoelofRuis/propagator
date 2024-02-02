package propagator

// Mutator collects and applies mutations from constraints
type Mutator struct {
	activeConstraintId constraintId
	mutations          []Mutation
	prevHead           int
	head               int
}

// newMutator Creates a new Mutator.
func newMutator() *Mutator {
	return &Mutator{
		activeConstraintId: -1,
		mutations:          make([]Mutation, 0, 10),
		prevHead:           0,
		head:               0,
	}
}

// Add adds mutations to be applied to the mutator without applying them.
func (m *Mutator) Add(updates ...Mutation) {
	for _, update := range updates {
		if update.domain == nil || len(update.indices) == 0 {
			continue
		}
		update.constraintId = m.activeConstraintId
		m.mutations = append(m.mutations, update)
	}
}

// setActiveConstraintId is called internally by the solver to notify the mutator of the constraint for which
// mutations are currently processed.
func (m *Mutator) setActiveConstraintId(c constraintId) {
	m.activeConstraintId = c
}

func (m *Mutator) apply() {
	m.prevHead = m.head
	for m.head < len(m.mutations) {
		m.mutations[m.head].apply()
		m.head++
	}
}

func (m *Mutator) revertAll() {
	for m.head > 0 {
		m.head--
		m.mutations[m.head].revert()
	}
	m.mutations = m.mutations[:0]
}

func (m *Mutator) revertPrevious() {
	for m.head > m.prevHead {
		m.head--
		m.mutations[m.head].revert()
		m.mutations = m.mutations[:m.head]
	}
}

// Mutation defines a mutation to the probability and priority set for the indices of a Domain.
type Mutation struct {
	domain      *Domain
	indices     []int
	probability Probability
	priority    Priority

	constraintId   constraintId
	reverseIndices []reverseIndex
}

// DoNothing is the update that changes nothing to a domain.
var DoNothing = Mutation{}

// reverseIndex stores an indexId and constraintId together with probability and priority so it can be reversed.
type reverseIndex struct {
	indexId      int
	constraintId constraintId
	oldProbPrio  packedProbPrio
}

// apply applies the changes defined by this mutation and tracks the changed indices, so they can be reverted.
func (u *Mutation) apply() {
	u.reverseIndices = make([]reverseIndex, 0, len(u.indices))
	for _, i := range u.indices {
		idxProbability := u.domain.model.domainIndexProbability[u.domain.id][i]
		if idxProbability < 10e-10 {
			continue
		}

		var currentProbPrio packedProbPrio
		if u.constraintId == -1 {
			currentProbPrio = u.domain.model.domainIndexDefaultModifiers[u.domain.id][i]
		} else {
			currentProbPrio = u.domain.model.domainIndexConstraintModifiers[u.domain.id][i][u.constraintId]
		}

		currentProbability, currentPriority := unpackPriorityProbability(currentProbPrio)

		shouldUpdateProbability := u.probability < currentProbability
		shouldUpdatePriority := u.priority > currentPriority

		if !shouldUpdateProbability && !shouldUpdatePriority {
			continue
		}

		adjustedProbability := Probability(1.0)
		adjustedPriority := Priority(0)

		if shouldUpdateProbability {
			adjustedProbability = u.probability
		}

		if shouldUpdatePriority {
			adjustedPriority = u.priority
		}

		adjustedProbPrio := packPriorityProbability(adjustedProbability, adjustedPriority)

		if u.constraintId == -1 {
			u.domain.model.domainIndexDefaultModifiers[u.domain.id][i] = adjustedProbPrio
		} else {
			u.domain.model.domainIndexConstraintModifiers[u.domain.id][i][u.constraintId] = adjustedProbPrio
		}
		u.reverseIndices = append(u.reverseIndices, reverseIndex{indexId: i, constraintId: u.constraintId, oldProbPrio: currentProbPrio})
	}

	if len(u.reverseIndices) > 0 {
		u.domain.update()
	}
}

// revert reverts the changes done by this mutation.
func (u *Mutation) revert() {
	if len(u.reverseIndices) == 0 {
		return
	}

	for _, revIdx := range u.reverseIndices {
		if u.constraintId == -1 {
			u.domain.model.domainIndexDefaultModifiers[u.domain.id][revIdx.indexId] = revIdx.oldProbPrio
		} else {
			u.domain.model.domainIndexConstraintModifiers[u.domain.id][revIdx.indexId][revIdx.constraintId] = revIdx.oldProbPrio
		}
	}

	u.domain.update()
}
