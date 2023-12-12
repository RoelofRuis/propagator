package propagator

import "math"

// Model holds the tracked variables and the constraints between them.
type Model struct {
	domainConstraints map[*Domain][]constraintId
	constraints       []boundConstraint
	domains           []*Domain // TODO: hopefully we can completely replace this with all inline arrays in the model. (only keep id as reference)

	// domain state
	domainNumIndices []int
	domainNames      []string
	domainEntropy    []float64
}

func (m *Model) update() {
	// select domain ids that need updating

	// foreach updatable domain Di
	// domainEntropy[Di] = math.Inf(+1)
}

func (m *Model) NumIndicesOf(d *Domain) int {
	return m.domainNumIndices[d.id]
}

func (m *Model) NameOf(d *Domain) string {
	return m.domainNames[d.id]
}

func (m *Model) EntropyOf(d *Domain) float64 {
	if !math.IsInf(m.domainEntropy[d.id], +1) {
		return m.domainEntropy[d.id]
	}

	if d.sumProbability == 0.0 {
		m.domainEntropy[d.id] = math.Inf(-1)
		return m.domainEntropy[d.id]
	}

	entropy := 0.0
	for _, idx := range d.indices {
		if idx.isBanned || idx.priority != d.minPriority {
			continue
		}
		weightedProb := idx.probability / d.sumProbability
		entropy += weightedProb * math.Log2(weightedProb)
	}
	m.domainEntropy[d.id] = -entropy
	return m.domainEntropy[d.id]
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

type boundConstraint struct {
	constraint    Constraint
	linkedDomains []*Domain
}

// ModelBuilder holds information about a model under construction.
type ModelBuilder struct {
	domains           []*Domain
	domainConstraints map[*Domain][]constraintId
	constraints       []boundConstraint
}

// BuildModel returns an empty model builder which can be used to build a constraint propagation model.
func BuildModel() *ModelBuilder {
	return &ModelBuilder{
		domains:           []*Domain{},
		domainConstraints: make(map[*Domain][]constraintId),
		constraints:       []boundConstraint{},
	}
}

// AddDomain adds a domain to the model that will be actively tracked and solved for.
// Domains not added via this function can still be modified via attached constraints.
func (m *ModelBuilder) AddDomain(domain *Domain) {
	domain.id = len(m.domains)
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
	domainNumIndices := make([]int, len(m.domains))
	domainNames := make([]string, len(m.domains))
	domainEntropy := make([]float64, len(m.domains))
	for i, domain := range m.domains {
		domainNumIndices[i] = len(domain.indices)
		domainNames[i] = domain.name
		domainEntropy[i] = math.Inf(+1)
	}

	return Model{
		domainConstraints: m.domainConstraints,
		constraints:       m.constraints,
		domains:           m.domains,
		domainNumIndices:  domainNumIndices,
		domainNames:       domainNames,
		domainEntropy:     domainEntropy,
	}
}
