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
	probability float64
	priority    int

	constraintId   constraintId
	reverseIndices []reverseIndex
}

// DoNothing is the update that changes nothing to a domain.
var DoNothing = Mutation{}

// reverseIndex stores an index id together with the previous index, so it can later be reversed.
type reverseIndex struct {
	id  int
	old *index
}

// apply applies the changes defined by this mutation and tracks the changed indices, so they can be reverted.
func (u *Mutation) apply() {
	u.reverseIndices = make([]reverseIndex, 0, len(u.indices))
	for _, i := range u.indices {
		oldIndex := u.domain.getIndex(i)
		newIndex, isUpdated := oldIndex.adjust(
			u.constraintId,
			u.probability,
			u.priority,
		)
		if !isUpdated {
			continue
		}

		u.reverseIndices = append(u.reverseIndices, reverseIndex{id: i, old: oldIndex})
		u.domain.setIndex(i, newIndex)
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
		u.domain.setIndex(revIdx.id, revIdx.old)
	}

	u.domain.update()
}
