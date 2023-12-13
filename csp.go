// Package propagator contains a framework for defining and solving constraint satisfaction problems.
package propagator

import (
	"math"
	"slices"
)

// Problem holds information about a constraint satisfaction problem under construction.
// Use NewProblem to start defining a new problem.
type Problem struct {
	model                  *Model
	domains                []*Domain
	nextDomainId           DomainId
	domainNames            []string
	domainIndices          [][]*index
	domainAvailableIndices [][]int
	domainConstraints      map[DomainId][]constraintId
	constraints            []boundConstraint
}

// NewProblem returns a builder with which to define a constraint satisfaction problem.
func NewProblem() *Problem {
	return &Problem{
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

// DomainValue represents the initialization data for a domain value.
type DomainValue[T comparable] struct {
	Priority    int
	Probability float64
	Value       T
}

// Model returns the initialized model.
// This should be called after the problem is completely defined.
func (c *Problem) Model() Model {
	numDomains := c.nextDomainId

	domainNumIndices := make([]int, numDomains)
	domainEntropy := make([]float64, numDomains)
	domainVersions := make([]int, numDomains)
	domainSumProbability := make([]float64, numDomains)
	domainMinPriority := make([]int, numDomains)

	for i := 0; i < numDomains; i++ {
		domainNumIndices[i] = len(c.domainIndices[i])
		domainEntropy[i] = math.Inf(+1)
		domainVersions[i] = 0
		domainSumProbability[i] = 0.0
		domainMinPriority[i] = 0
	}

	c.model.Domains = c.domains
	c.model.domainConstraints = c.domainConstraints
	c.model.constraints = c.constraints
	c.model.domainNumIndices = domainNumIndices
	c.model.domainNames = c.domainNames
	c.model.domainEntropy = domainEntropy
	c.model.domainVersions = domainVersions
	c.model.domainSumProbability = domainSumProbability
	c.model.domainMinPriority = domainMinPriority
	c.model.domainIndices = c.domainIndices
	c.model.domainAvailableIndices = c.domainAvailableIndices
	c.model.indexBuffer = make([]int, slices.Max(domainNumIndices))

	for _, domain := range c.domains {
		domain.update()
	}

	return *c.model
}

// AddConstraint adds a constraint to the Problem definition.
func (c *Problem) AddConstraint(constraint Constraint) {
	index := len(c.constraints)
	domainsInScope := constraint.Scope()
	if len(domainsInScope) == 0 {
		panic("constraint scope contains no Domains")
	}

	c.constraints = append(c.constraints, boundConstraint{constraint, domainsInScope})
	for _, domainInScope := range domainsInScope {
		constraintLinks := c.domainConstraints[domainInScope]
		constraintLinks = append(constraintLinks, index)
		c.domainConstraints[domainInScope] = constraintLinks
	}
}

// AddVariable adds a variable to the Problem definition.
func AddVariable[T comparable](csp *Problem, name string, initialValues []DomainValue[T]) *Variable[T] {
	values := make([]T, len(initialValues))
	indices := make([]*index, len(initialValues))

	for idx, value := range initialValues {
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
		values[idx] = value.Value
	}

	domain := Domain{
		id:    csp.nextDomainId,
		model: csp.model,
	}

	variable := &Variable[T]{
		Domain:             domain,
		values:             values,
		cachedValueVersion: 0,
		cachedValues:       make([]T, 0, len(initialValues)),
	}

	csp.nextDomainId++
	csp.domains = append(csp.domains, &domain)
	csp.domainNames = append(csp.domainNames, name)
	csp.domainIndices = append(csp.domainIndices, indices)
	csp.domainAvailableIndices = append(csp.domainAvailableIndices, make([]int, 0, len(indices)))

	return variable
}

// AddVariableFromValues adds a variable to the Problem definition and automatically gives all values equal probability
// and priority.
func AddVariableFromValues[T comparable](csp *Problem, name string, values []T) *Variable[T] {
	domainValues := make([]DomainValue[T], len(values))
	for i, value := range values {
		domainValues[i] = DomainValue[T]{0, 1.0, value}
	}
	return AddVariable[T](csp, name, domainValues)
}
