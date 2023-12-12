package propagator

type Variable[T comparable] struct {
	Domain
	values          []T
	availableValues []T
}

func (v *Variable[T]) AllowedValues() []T {
	v.availableValues = v.availableValues[:0]
	for _, idx := range v.availableIndices() {
		v.availableValues = append(v.availableValues, v.values[idx])
	}
	return v.availableValues
}

func (v *Variable[T]) IsValueAllowed(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

func (v *Variable[T]) Exists(check func(a T) bool) bool {
	for _, availableValue := range v.AllowedValues() {
		if check(availableValue) {
			return true
		}
	}
	return false
}

func (v *Variable[T]) ForEach(check func(a T) bool) bool {
	for _, availableValue := range v.AllowedValues() {
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
	v.indexBuffer = v.indexBuffer[:0]
	for _, availableIndex := range v.availableIndices() {
		if shouldBan(v.values[availableIndex]) {
			v.indexBuffer = append(v.indexBuffer, availableIndex)
		}
	}
	return v.Exclude(v.indexBuffer...)
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

func IdsOf[T comparable](vars ...*Variable[T]) []DomainId {
	domainIds := make([]DomainId, 0, len(vars))
	for _, v := range vars {
		domainIds = append(domainIds, v.Domain.id)
	}
	return domainIds
}
