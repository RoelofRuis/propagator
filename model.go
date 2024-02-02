package propagator

import (
	"fmt"
	"strings"
)

// DomainId is a reference to a domain.
type DomainId = int

// Model holds the tracked variables and the constraints between them.
type Model struct {
	// Domains allow access to the indices associated with the variables of this Model.
	Domains []*Domain

	// domainConstraints allows to look up the constraints that apply to a particular domain.
	domainConstraints map[DomainId][]constraintId
	// constraints holds all constraints indexed by their constraintId.
	constraints []boundConstraint

	// Various slices of data describing the domain states in this Model.
	domainHidden           []bool
	domainNumIndices       []int
	domainNames            []string
	domainEntropy          []float64
	domainVersions         []int
	domainSumProbability   []Probability
	domainMinPriority      []Priority
	domainIndexProbability [][]Probability
	domainIndexPriority    [][]Priority

	// deprecated
	domainIndexDefaultModifiers [][]packedProbPrio

	// deprecated
	domainIndexConstraintModifiers [][][]packedProbPrio

	domainAvailableIndices [][]int

	indexBuffer []int
}

func (m *Model) String() string {
	b := strings.Builder{}
	for _, domain := range m.Domains {
		name := m.domainNames[domain.id]
		prob := m.domainIndexProbability[domain.id]
		prio := m.domainIndexPriority[domain.id]
		mods := m.domainIndexConstraintModifiers[domain.id]
		b.WriteString(fmt.Sprintf("%s %v %v\n%v\n", name, prob, prio, mods))
	}
	return b.String()
}

// boundConstraint defines the link between a Constraint and its related Domains.
type boundConstraint struct {
	constraint    Constraint
	linkedDomains []DomainId
}

// IsSolved returns whether this model currently is in a solved state.
func (m *Model) IsSolved() bool {
	for _, domain := range m.Domains {
		if domain.IsHidden() {
			continue
		}
		if !domain.IsAssigned() {
			return false
		}
	}
	return true
}
