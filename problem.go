// Package propagator contains a framework for defining and solving constraint satisfaction problems.
package propagator

import (
	"math"
	"slices"
)

// Problem holds information about a constraint satisfaction problem under construction.
// Use NewProblem to start defining a new problem.
type Problem struct {
	model                          *Model
	domains                        []*Domain
	nextDomainId                   DomainId
	domainNames                    []string
	domainHidden                   []bool
	domainIndexConstraintModifiers [][][]packedProbPrio
	domainIndexDefaultModifiers    [][]packedProbPrio
	domainIndexProbability         [][]Probability
	domainIndexPriority            [][]Priority
	domainAvailableIndices         [][]int
	domainConstraints              map[DomainId][]constraintId
	constraints                    []boundConstraint
}

// NewProblem returns a builder with which to define a constraint satisfaction problem.
func NewProblem() *Problem {
	return &Problem{
		model:                          &Model{},
		domains:                        []*Domain{},
		nextDomainId:                   0,
		domainNames:                    []string{},
		domainHidden:                   []bool{},
		domainConstraints:              make(map[DomainId][]constraintId),
		domainIndexConstraintModifiers: [][][]packedProbPrio{},
		domainIndexDefaultModifiers:    [][]packedProbPrio{},
		domainAvailableIndices:         [][]int{},
		constraints:                    []boundConstraint{},
	}
}

// DomainValue represents the initialization data for a domain value.
type DomainValue[T comparable] struct {
	Priority    Priority
	Probability Probability
	Value       T
}

// Model returns the initialized model.
// This should be called after the problem is completely defined.
func (c *Problem) Model() Model {
	numDomains := c.nextDomainId

	domainNumIndices := make([]int, numDomains)
	domainEntropy := make([]float64, numDomains)
	domainVersions := make([]int, numDomains)
	domainSumProbability := make([]Probability, numDomains)
	domainMinPriority := make([]Priority, numDomains)

	for i := 0; i < numDomains; i++ {
		domainNumIndices[i] = len(c.domainIndexDefaultModifiers[i])
		domainEntropy[i] = math.Inf(+1)
		domainVersions[i] = 0
		domainSumProbability[i] = 0.0
		domainMinPriority[i] = 0

		for j := 0; j < domainNumIndices[i]; j++ {
			modifiers := make([]packedProbPrio, len(c.constraints))
			for k := 0; k < len(c.constraints); k++ {
				modifiers[k] = packPriorityProbability(1.0, 0)
			}
			c.domainIndexConstraintModifiers[i][j] = modifiers
		}
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
	c.model.domainIndexConstraintModifiers = c.domainIndexConstraintModifiers
	c.model.domainIndexDefaultModifiers = c.domainIndexDefaultModifiers
	c.model.domainIndexProbability = c.domainIndexProbability
	c.model.domainIndexPriority = c.domainIndexPriority
	c.model.domainHidden = c.domainHidden
	c.model.domainAvailableIndices = c.domainAvailableIndices
	c.model.indexBuffer = make([]int, 0, slices.Max(domainNumIndices))

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
	return newVariable(csp, name, initialValues, false)
}

// AddHiddenVariable adds a hidden variable to the Problem definition.
func AddHiddenVariable[T comparable](csp *Problem, name string, initialValues []DomainValue[T]) *Variable[T] {
	return newVariable(csp, name, initialValues, true)
}

// AddVariableFromValues adds a variable to the Problem definition and automatically gives all values equal probability
// and priority.
func AddVariableFromValues[T comparable](csp *Problem, name string, values []T) *Variable[T] {
	domainValues := make([]DomainValue[T], len(values))
	for i, value := range values {
		domainValues[i] = DomainValue[T]{0, 1.0, value}
	}
	return newVariable[T](csp, name, domainValues, false)
}

// AddHiddenVariableFromValues adds a hidden variable to the Problem definition and automatically gives all values
// equal probability and priority.
func AddHiddenVariableFromValues[T comparable](csp *Problem, name string, values []T) *Variable[T] {
	domainValues := make([]DomainValue[T], len(values))
	for i, value := range values {
		domainValues[i] = DomainValue[T]{0, 1.0, value}
	}
	return newVariable[T](csp, name, domainValues, true)
}

// newVariable builds a new variable definition bound to the given problem.
func newVariable[T comparable](csp *Problem, name string, initialValues []DomainValue[T], hidden bool) *Variable[T] {
	values := make([]T, len(initialValues))
	indexConstraintModifiers := make([][]packedProbPrio, len(initialValues))
	indexDefaultModifiers := make([]packedProbPrio, len(initialValues))
	indexDefaultProbability := make([]Probability, len(initialValues))
	indexDefaultPriority := make([]Priority, len(initialValues))

	for idx, value := range initialValues {
		indexDefaultProbability[idx] = value.Probability
		indexDefaultModifiers[idx] = packPriorityProbability(value.Probability, value.Priority)
		indexDefaultPriority[idx] = value.Priority
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
	csp.domainIndexProbability = append(csp.domainIndexProbability, indexDefaultProbability)
	csp.domainIndexPriority = append(csp.domainIndexPriority, indexDefaultPriority)
	csp.domainIndexConstraintModifiers = append(csp.domainIndexConstraintModifiers, indexConstraintModifiers)
	csp.domainIndexDefaultModifiers = append(csp.domainIndexDefaultModifiers, indexDefaultModifiers)
	csp.domainAvailableIndices = append(csp.domainAvailableIndices, make([]int, 0, len(initialValues)))
	csp.domainHidden = append(csp.domainHidden, hidden)

	return variable
}
