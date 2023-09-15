package propagator

// Mutator collects and applies mutations from constraints
type Mutator struct {
	activeConstraintId constraintId
	mutations          []Mutation
	prevHead           int
	head               int
}

// NewMutator Creates a new Mutator.
func NewMutator() *Mutator {
	return &Mutator{
		activeConstraintId: -1,
		prevHead:           0,
		head:               0,
	}
}

// setActiveConstraintId is called internally by the solver to notify the mutator of the constraint for which
// mutations are currently processed.
func (m *Mutator) setActiveConstraintId(c constraintId) {
	m.activeConstraintId = c
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
	m.mutations = []Mutation{}
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
	probability float64
	priority    int

	constraintId   constraintId
	reverseIndices map[int]index
}

// DoNothing is the update that changes nothing to a domain.
var DoNothing = Mutation{}

// apply applies the changes defined by this mutation and tracks the changed indices, so they can be reverted.
func (u *Mutation) apply() {
	u.reverseIndices = make(map[int]index)
	for _, i := range u.indices {
		newIndex, isUpdated := u.domain.indices[i].adjust(
			u.constraintId,
			u.probability,
			u.priority,
		)
		if !isUpdated {
			continue
		}

		u.reverseIndices[i] = u.domain.indices[i]
		u.domain.indices[i] = newIndex
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

	for i, oldIndex := range u.reverseIndices {
		u.domain.indices[i] = oldIndex
	}

	u.domain.update()
}
