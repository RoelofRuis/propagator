package propagator

// DomainId is a reference to a domain.
type DomainId = int

// Model holds the tracked variables and the constraints between them.
type Model struct {
	// Domains allow access to the indices associated with the variables of this Model.
	Domains []*Domain

	domainConstraints map[DomainId][]constraintId
	constraints       []boundConstraint

	domainNumIndices       []int
	domainNames            []string
	domainEntropy          []float64
	domainVersions         []int
	domainSumProbability   []float64
	domainMinPriority      []int
	domainIndices          [][]*index
	domainAvailableIndices [][]int

	indexBuffer []int
}

// boundConstraint defines the link between a Constraint and its related Domains.
type boundConstraint struct {
	constraint    Constraint
	linkedDomains []DomainId
}

// IsSolved returns whether this model currently is in a solved state.
func (m *Model) IsSolved() bool {
	for _, domain := range m.Domains {
		if !domain.IsAssigned() {
			return false
		}
	}
	return true
}
