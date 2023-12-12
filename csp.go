package propagator

import "math"

// CSP holds information about a constraint satisfaction problem under construction.
type CSP struct {
	model                  *Model
	domains                []*Domain
	nextDomainId           DomainId
	domainNames            []string
	domainIndices          [][]*index
	domainAvailableIndices [][]int
	domainConstraints      map[DomainId][]constraintId
	constraints            []boundConstraint
}

// NewCSP returns a builder with which to define a constraint satisfaction problem.
func NewCSP() *CSP {
	return &CSP{
		model:                  &Model{},
		domains:                []*Domain{},
		nextDomainId:           0,
		domainNames:            []string{},
		domainConstraints:      make(map[DomainId][]constraintId),
		domainIndices:          [][]*index{},
		domainAvailableIndices: [][]int{},
		constraints:            []boundConstraint{},
	}
}

func AddVariableFromValues[T comparable](csp *CSP, name string, values []T) *Variable[T] {
	return AddVariable[T](csp, name, AsDomainValues(values...))
}

func AddVariable[T comparable](csp *CSP, name string, initialValues []DomainValue[T]) *Variable[T] {
	values := make([]T, len(initialValues))
	indices := make([]*index, len(initialValues))

	for idx, value := range initialValues {
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
		values[idx] = value.Value
	}

	domain := Domain{
		id:          csp.nextDomainId,
		model:       csp.model,
		indexBuffer: make([]int, 0, len(initialValues)),
	}

	variable := &Variable[T]{
		Domain:          domain,
		values:          values,
		availableValues: make([]T, 0, len(initialValues)),
	}

	csp.nextDomainId++
	csp.domains = append(csp.domains, &domain)
	csp.domainNames = append(csp.domainNames, name)
	csp.domainIndices = append(csp.domainIndices, indices)
	csp.domainAvailableIndices = append(csp.domainAvailableIndices, make([]int, 0, len(indices)))

	return variable
}

// AddConstraint adds a constraint to the model.
func (m *CSP) AddConstraint(constraint Constraint) {
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

// GetModel returns the initialized model. Should be called after the problem is completely defined.
func (m *CSP) GetModel() Model {
	numDomains := m.nextDomainId

	domainNumIndices := make([]int, numDomains)
	domainEntropy := make([]float64, numDomains)
	domainVersions := make([]int, numDomains)
	domainSumProbability := make([]float64, numDomains)
	domainMinPriority := make([]int, numDomains)

	for i := 0; i < numDomains; i++ {
		domainNumIndices[i] = len(m.domainIndices[i])
		domainEntropy[i] = math.Inf(+1)
		domainVersions[i] = 0
		domainSumProbability[i] = 0.0
		domainMinPriority[i] = 0
	}

	m.model.domains = m.domains
	m.model.domainConstraints = m.domainConstraints
	m.model.constraints = m.constraints
	m.model.domainNumIndices = domainNumIndices
	m.model.domainNames = m.domainNames
	m.model.domainEntropy = domainEntropy
	m.model.domainVersions = domainVersions
	m.model.domainSumProbability = domainSumProbability
	m.model.domainMinPriority = domainMinPriority
	m.model.domainIndices = m.domainIndices
	m.model.domainAvailableIndices = m.domainAvailableIndices

	for _, domain := range m.domains {
		domain.update()
	}

	return *m.model
}
