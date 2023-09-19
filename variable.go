package propagator

import (
	"fmt"
	"strings"
)

// Variable extends Domain to provide also the specifically typed values belonging to the domain indices.
type Variable[T comparable] struct {
	*Domain
	// values holds the values associated with the indices of this variable.
	values []T
	// exclusionBuffer holds a pre-allocated buffer storing indices collected through ExcludeBy.
	exclusionBuffer []int
	// availableValueBuffer holds a pre-allocated buffer storing the list of available values.
	availableValueBuffer []T
	// availableValuesVersion holds the domain version on which the availableValueBuffer was calculated.
	availableValuesVersion int
}

// NewVariable instantiates a new variable from a name and a given set of domain values.
// The domain values allow for specifying the initial probability and the priority.
func NewVariable[T comparable](name string, initialValues []DomainValue[T]) *Variable[T] {
	indices := make([]*index, len(initialValues))
	values := make([]T, len(initialValues))

	for idx, value := range initialValues {
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
		values[idx] = value.Value
	}

	return &Variable[T]{
		Domain:                 NewDomain(name, indices),
		values:                 values,
		exclusionBuffer:        make([]int, 0, len(initialValues)),
		availableValueBuffer:   make([]T, 0, len(initialValues)),
		availableValuesVersion: 0,
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

// AllowedValues returns all the non-excluded values the variable currently holds.
func (v Variable[T]) AllowedValues() []T {
	if v.availableValuesVersion < v.Domain.version {
		v.availableValueBuffer = v.availableValueBuffer[:0]
		for _, availableIndex := range v.Domain.availableIndices {
			v.availableValueBuffer = append(v.availableValueBuffer, v.values[availableIndex])
		}
		v.availableValuesVersion = v.Domain.version
	}

	return v.availableValueBuffer
}

// IsValueAllowed returns whether the given value still remains free as a value to be selected.
func (v Variable[T]) IsValueAllowed(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

// Exists checks whether there exists a non-banned value that passes the provided check.
func (v Variable[T]) Exists(check func(a T) bool) bool {
	for _, availableIndex := range v.availableIndices {
		if check(v.values[availableIndex]) {
			return true
		}
	}
	return false
}

// ForEach checks whether all non-banned values pass the provided check.
func (v Variable[T]) ForEach(check func(a T) bool) bool {
	for _, availableIndex := range v.availableIndices {
		if !check(v.values[availableIndex]) {
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

// GetAssignedValue returns the value of the fixed index.
// It panics if the variable is not fixed: use IsAssigned to check for this.
func (v Variable[T]) GetAssignedValue() T {
	idx := v.GetAssignedIndex()
	if idx == -1 {
		panic("Trying to call GetAssignedValue on non fixed variable. Use IsAssigned to check.")
	}
	return v.values[idx]
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
	for _, availableIndex := range v.Domain.availableIndices {
		if v.values[availableIndex] == value {
			return v.Update(probabilityFactor, priority, availableIndex)
		}
	}
	return DoNothing
}

// AssignByValue returns the Mutation that fixes this variable to the given value.
func (v Variable[T]) AssignByValue(value T) Mutation {
	for _, availableIndex := range v.availableIndices {
		if value == v.values[availableIndex] {
			return v.Assign(availableIndex)
		}
	}
	return DoNothing
}

// ExcludeBy returns the Mutation that bans all values for which shouldBan evaluates to true.
func (v Variable[T]) ExcludeBy(shouldBan func(T) bool) Mutation {
	v.exclusionBuffer = v.exclusionBuffer[:0]
	for _, availableIndex := range v.Domain.availableIndices {
		if shouldBan(v.values[availableIndex]) {
			v.exclusionBuffer = append(v.exclusionBuffer, availableIndex)
		}
	}
	return v.Exclude(v.exclusionBuffer...)
}

// ExcludeByValue returns the Mutation that bans all the given values.
func (v Variable[T]) ExcludeByValue(values ...T) Mutation {
	return v.ExcludeBy(func(a T) bool {
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
	for index, s := range v.values {
		value := s
		prob := v.Domain.indices[index].probability
		prio := v.Domain.indices[index].priority
		banned := true
		for _, availableIndex := range v.Domain.availableIndices {
			if index == availableIndex {
				banned = false
				break
			}
		}
		if banned {
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
