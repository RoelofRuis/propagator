package propagator

type Variable[T comparable] struct {
	Domain
	values []T

	cachedValueVersion int
	cachedValues       []T
}

func (v *Variable[T]) AvailableValues() []T {
	if v.version() != v.cachedValueVersion {
		availableIndices := v.availableIndices()
		v.cachedValues = v.cachedValues[:len(availableIndices)]
		for i, idx := range availableIndices {
			v.cachedValues[i] = v.values[idx]
		}
	}
	return v.cachedValues
}

func (v *Variable[T]) IsValueAvailable(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

func (v *Variable[T]) Exists(check func(a T) bool) bool {
	for _, availableValue := range v.AvailableValues() {
		if check(availableValue) {
			return true
		}
	}
	return false
}

func (v *Variable[T]) ForEach(check func(a T) bool) bool {
	for _, availableValue := range v.AvailableValues() {
		if !check(availableValue) {
			return false
		}
	}
	return true
}

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

func (v *Variable[T]) GetAssignedValue() T {
	if v.IsAssigned() {
		return v.values[v.availableIndices()[0]]
	}
	panic("Trying to GetAssignedValue on non fixed variable. Use IsAssigned to check.")
}

func (v *Variable[T]) UpdatePriorityByValue(priority int, value T) Mutation {
	return v.UpdateByValue(1.0, priority, value)
}

func (v *Variable[T]) UpdateProbabilityByValue(factory float64, value T) Mutation {
	return v.UpdateByValue(factory, 0, value)
}

func (v *Variable[T]) UpdateByValue(probabilityFactor float64, priority int, value T) Mutation {
	for _, availableIndex := range v.availableIndices() {
		if v.values[availableIndex] == value {
			return v.Update(probabilityFactor, priority, availableIndex)
		}
	}
	return DoNothing
}

func (v *Variable[T]) AssignByValue(value T) Mutation {
	for _, availableIndex := range v.availableIndices() {
		if v.values[availableIndex] == value {
			return v.Assign(availableIndex)
		}
	}
	return DoNothing
}

func (v *Variable[T]) ExcludeBy(shouldBan func(T) bool) Mutation {
	v.model.indexBuffer = v.model.indexBuffer[:0]
	for _, availableIndex := range v.availableIndices() {
		if shouldBan(v.values[availableIndex]) {
			v.model.indexBuffer = append(v.model.indexBuffer, availableIndex)
		}
	}
	return v.Exclude(v.model.indexBuffer...)
}

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

func IdsOf[T comparable](vars ...*Variable[T]) []DomainId {
	domainIds := make([]DomainId, 0, len(vars))
	for _, v := range vars {
		domainIds = append(domainIds, v.Domain.id)
	}
	return domainIds
}
