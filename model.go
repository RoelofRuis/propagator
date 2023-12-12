package propagator

type DomainId = int

// Model holds the tracked variables and the constraints between them.
type Model struct {
	domainConstraints map[DomainId][]constraintId // TODO: Replace index with DomainId?
	constraints       []boundConstraint
	domains           []*Domain

	domainNumIndices       []int
	domainNames            []string
	domainEntropy          []float64
	domainVersions         []int
	domainSumProbability   []float64
	domainMinPriority      []int
	domainIndices          [][]*index
	domainAvailableIndices [][]int
}

// boundConstraint defines the link between a Constraint and its related domains.
type boundConstraint struct {
	constraint    Constraint
	linkedDomains []DomainId
}

// IsSolved returns whether this model currently is in a solved state.
func (m *Model) IsSolved() bool {
	for _, domain := range m.domains {
		if !domain.IsAssigned() {
			return false
		}
	}
	return true
}
