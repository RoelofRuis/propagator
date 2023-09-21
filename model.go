package propagator

// Constraint describes the way domains depend on each other and allows for propagating updated values.
type Constraint interface {
	// Scope returns all domains that are influenced by this constraint.
	Scope() []Domain2
	// Propagate is called every time a domain in the constraint scope is updated and allows for further updates to be passed.
	Propagate(m *Mutator)
}

// Model holds the tracked variables and the constraints between them.
type Model struct {
	domainConstraints map[Domain2][]constraintId
	constraints       []boundConstraint
	domains           []Domain2
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

type constraintId = int

type boundConstraint struct {
	constraint    Constraint
	linkedDomains []Domain2
}

// ModelBuilder holds information about a model under construction.
type ModelBuilder struct {
	domainIndex       int
	domains           []Domain2
	domainConstraints map[Domain2][]constraintId
	constraints       []boundConstraint
}

// BuildModel returns an empty model builder which can be used to build a constraint propagation model.
func BuildModel() *ModelBuilder {
	return &ModelBuilder{
		domains:           []Domain2{},
		domainConstraints: make(map[Domain2][]constraintId),
		constraints:       []boundConstraint{},
	}
}

// AddDomain adds a domain to the model that will be actively tracked and solved for.
// Domains not added via this function can still be modified via attached constraints.
func (m *ModelBuilder) AddDomain(domain Domain2) {
	domain.setId(m.domainIndex)
	m.domainIndex++
	m.domains = append(m.domains, domain)
}

// AddConstraint adds a constraint to the model.
func (m *ModelBuilder) AddConstraint(constraint Constraint) {
	index := len(m.constraints)
	domainsInScope := constraint.Scope()
	if len(domainsInScope) == 0 {
		panic("constraint scope contains no domains")
	}

	m.constraints = append(m.constraints, boundConstraint{constraint, domainsInScope})
	for _, domainInScope := range domainsInScope {
		constraintLinks := m.domainConstraints[domainInScope]
		constraintLinks = append(constraintLinks, index)
		m.domainConstraints[domainInScope] = constraintLinks
	}
}

// Build returns the completely initialized model.
func (m *ModelBuilder) Build() Model {
	return Model{
		domains:           m.domains,
		domainConstraints: m.domainConstraints,
		constraints:       m.constraints,
	}
}
