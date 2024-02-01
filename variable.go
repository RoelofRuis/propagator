package propagator

// Variable holds some number of values of type T.
// It is an expansion of Domain that associates the actual types relevant to the problem with the probability and
// priority of the indices.
type Variable[T comparable] struct {
	Domain
	values []T

	// cachedValueVersion is the Domain.version for which the cachedValues are calculated.
	cachedValueVersion int
	cachedValues       []T
}

// AvailableValues returns the list of still allowed values.
func (v *Variable[T]) AvailableValues() []T {
	if v.version() != v.cachedValueVersion {
		availableIndices := v.AvailableIndices()
		v.cachedValues = v.cachedValues[:len(availableIndices)]
		for i, idx := range availableIndices {
			v.cachedValues[i] = v.values[idx]
		}
	}
	return v.cachedValues
}

// AvailableIndicesAndValues returns two slices of equal length, the first containing the still available indices and
// the second containing the corresponding values.
func (v *Variable[T]) AvailableIndicesAndValues() ([]int, []T) {
	return v.AvailableIndices(), v.AvailableValues()
}

// IsValueAvailable checks whether a given value is still allowed to be selected.
func (v *Variable[T]) IsValueAvailable(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

// Exists checks whether within the available values there exists a value passing the provided check.
func (v *Variable[T]) Exists(check func(a T) bool) bool {
	for _, availableValue := range v.AvailableValues() {
		if check(availableValue) {
			return true
		}
	}
	return false
}

// ForEach checks whether within the available values all values pass the provided check.
func (v *Variable[T]) ForEach(check func(a T) bool) bool {
	for _, availableValue := range v.AvailableValues() {
		if !check(availableValue) {
			return false
		}
	}
	return true
}

// HasAnyOf checks whether any of the given values are still allowed to be selected.
func (v *Variable[T]) HasAnyOf(values ...T) bool {
	return v.Exists(func(a T) bool {
		for _, value := range values {
			if a == value {
				return true
			}
		}
		return false
	})
}

// GetAssignedValue returns the single assigned value.
// This function panics if the variable is not assigned a single value. Use Domain.IsAssigned to check for this.
func (v *Variable[T]) GetAssignedValue() T {
	if v.IsAssigned() {
		return v.values[v.AvailableIndices()[0]]
	}
	panic("Trying to GetAssignedValue on non fixed variable. Use IsAssigned to check.")
}

// UpdatePriorityByValue creates a Mutation that updates the priority of the index associated with the given value.
func (v *Variable[T]) UpdatePriorityByValue(priority Priority, value T) Mutation {
	return v.UpdateByValue(1.0, priority, value)
}

// UpdateProbabilityByValue creates a Mutation that updates the probability of the index associated with the given value.
func (v *Variable[T]) UpdateProbabilityByValue(factory Probability, value T) Mutation {
	return v.UpdateByValue(factory, 0, value)
}

// UpdateByValue creates a Mutation that updates the probability and priority of the index associated with the given value.
func (v *Variable[T]) UpdateByValue(factor Probability, priority Priority, value T) Mutation {
	for _, availableIndex := range v.AvailableIndices() {
		if v.values[availableIndex] == value {
			return v.Update(factor, priority, availableIndex)
		}
	}
	return DoNothing
}

// AssignByValue creates a Mutation that assigns this variable the value T.
// The specific value must exist in the variable to be assignable.
func (v *Variable[T]) AssignByValue(value T) Mutation {
	for _, availableIndex := range v.AvailableIndices() {
		if v.values[availableIndex] == value {
			return v.Assign(availableIndex)
		}
	}
	return DoNothing
}

// ExcludeBy creates a Mutation that excludes all values for which shouldBan returns true.
func (v *Variable[T]) ExcludeBy(shouldBan func(T) bool) Mutation {
	v.model.indexBuffer = v.model.indexBuffer[:0]
	for _, availableIndex := range v.AvailableIndices() {
		if shouldBan(v.values[availableIndex]) {
			v.model.indexBuffer = append(v.model.indexBuffer, availableIndex)
		}
	}
	return v.Exclude(v.model.indexBuffer...)
}

// ExcludeByValue creates a Mutation that excludes all given values.
func (v *Variable[T]) ExcludeByValue(values ...T) Mutation {
	return v.ExcludeBy(func(a T) bool {
		for _, value := range values {
			if value == a {
				return true
			}
		}
		return false
	})
}
