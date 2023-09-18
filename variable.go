package propagator

import (
	"fmt"
	"strings"
)

// Variable extends Domain to provide also the specifically typed values belonging to the domain indices.
type Variable[T comparable] struct {
	*Domain
	states []State[T]
}

// State associates an index with a value.
type State[T comparable] struct {
	Index int
	Value T
}

// NewVariable instantiates a new variable from a name and a given set of domain values.
// The domain values allow for specifying the initial probability and the priority.
func NewVariable[T comparable](name string, values []DomainValue[T]) *Variable[T] {
	states := make([]State[T], len(values))
	indices := make([]*index, len(values))

	for idx, value := range values {
		states[idx] = State[T]{idx, value.Value}
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
	}

	return &Variable[T]{
		Domain: NewDomain(name, indices),
		states: states,
	}
}

// NewVariableFromValues instantiates a new variable from a name and a given set of values.
// All values will be given an equal probability and a default priority of 0.
func NewVariableFromValues[T comparable](name string, values []T) *Variable[T] {
	return NewVariable[T](name, AsDomainValues(values...))
}

// DomainsOf extracts the domains from the given list of variables.
func DomainsOf[T comparable](variables []*Variable[T]) []*Domain {
	domains := make([]*Domain, len(variables))
	for i, v := range variables {
		domains[i] = v.Domain
	}
	return domains
}

// AvailableStates returns all the non-banned states the variable currently holds.
func (v Variable[T]) AvailableStates() []State[T] {
	states := make([]State[T], v.Domain.availableIndexCount)
	i := 0
	for _, state := range v.states {
		if !v.Domain.IndexIsBanned(state.Index) {
			states[i] = state
			i++
		}
	}
	return states
}

// AvailableValues returns all the non-banned values the variable currently holds.
func (v Variable[T]) AvailableValues() []T {
	var values []T
	for _, state := range v.AvailableStates() {
		values = append(values, state.Value)
	}
	return values
}

// RemainsFree returns whether the given value still remains free as a value to be selected.
func (v Variable[T]) RemainsFree(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

// Exists checks whether there exists a non-banned value that passes the provided check.
func (v Variable[T]) Exists(check func(a T) bool) bool {
	for _, state := range v.AvailableStates() {
		if check(state.Value) {
			return true
		}
	}
	return false
}

// ForEach checks whether all non-banned values pass the provided check.
func (v Variable[T]) ForEach(check func(a T) bool) bool {
	for _, state := range v.AvailableStates() {
		if !check(state.Value) {
			return false
		}
	}
	return true
}

// HasAnyOf checks whether any of the given values are in the non-banned values.
func (v Variable[T]) HasAnyOf(values ...T) bool {
	return v.Exists(func(a T) bool {
		for _, value := range values {
			if a == value {
				return true
			}
		}
		return false
	})
}

// GetFixedValue returns the value of the fixed index.
// It panics if the variable is not fixed: use IsFixed to check for this.
func (v Variable[T]) GetFixedValue() T {
	idx := v.GetFixedIndex()
	if idx == -1 {
		panic("Trying to call GetFixedValue on non fixed variable. Use IsFixed to check.")
	}
	return v.states[idx].Value
}

// UpdatePriorityByValue returns the Mutation that adjusts the priority of the given value.
func (v Variable[T]) UpdatePriorityByValue(priority int, value T) Mutation {
	return v.UpdateByValue(1.0, priority, value)
}

// UpdateProbabilityByValue returns the Mutation that adjusts the probability of the given value.
func (v Variable[T]) UpdateProbabilityByValue(factor float64, value T) Mutation {
	return v.UpdateByValue(factor, 0, value)
}

// UpdateByValue returns teh Mutation that adjusts the probability and priority of the given value.
func (v Variable[T]) UpdateByValue(probabilityFactor float64, priority int, value T) Mutation {
	for _, s := range v.AvailableStates() {
		if value == s.Value {
			return v.Update(probabilityFactor, priority, s.Index)
		}
	}
	return DoNothing
}

// FixByValue returns the Mutation that fixes this variable to the given value.
func (v Variable[T]) FixByValue(value T) Mutation {
	for _, s := range v.AvailableStates() {
		if value == s.Value {
			return v.Fix(s.Index)
		}
	}
	return DoNothing
}

// BanBy returns the Mutation that bans all values for which shouldBan evaluates to true.
func (v Variable[T]) BanBy(shouldBan func(T) bool) Mutation {
	var indicesToBan []int
	for _, s := range v.AvailableStates() {
		if shouldBan(s.Value) {
			indicesToBan = append(indicesToBan, s.Index)
		}
	}
	return v.Ban(indicesToBan...)
}

// BanByValue returns the Mutation that bans all the given values.
func (v Variable[T]) BanByValue(values ...T) Mutation {
	return v.BanBy(func(a T) bool {
		for _, value := range values {
			if value == a {
				return true
			}
		}
		return false
	})
}

func (v Variable[T]) String() string {
	var str []string
	for _, s := range v.states {
		index := s.Index
		value := s.Value
		prob := v.Domain.indices[s.Index].probability
		prio := v.Domain.indices[s.Index].priority
		if v.Domain.IndexIsBanned(s.Index) {
			str = append(str, fmt.Sprintf("[#%d ✘] (P=%.3f) (%d) %v", index, prob, prio, value))
		} else {
			str = append(str, fmt.Sprintf("[#%d ✔] (P=%.3f) (%d) %v", index, prob, prio, value))
		}
	}
	return fmt.Sprintf("\nVAR [%s]\n%s", v.Name, strings.Join(str, "\n"))
}

// AsDomainValues instantiates default domain values from a list of values.
func AsDomainValues[T comparable](values ...T) []DomainValue[T] {
	domainValues := make([]DomainValue[T], len(values))
	for i, value := range values {
		domainValues[i] = DomainValue[T]{0, 1.0, value}
	}
	return domainValues
}

// DomainValue represents the initialization data for a domain value.
type DomainValue[T comparable] struct {
	Priority    int
	Probability float64
	Value       T
}
