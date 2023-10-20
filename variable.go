package propagator

import (
	"math"
)

type Domain interface {
	numIndices() int
	getIndex(i int) *index
	setIndex(i int, idx *index)
	update()
	getMinPriority() int
	getVersion() int

	GetName() string
	IsAssigned() bool
	IsUnassigned() bool
	IsInContradiction() bool
	Entropy() float64
	Assign(i int) Mutation
	Exclude(i ...int) Mutation
}

type Variable[T comparable] struct {
	id               int
	name             string
	indices          []*index
	values           []T
	availableIndices []int
	availableValues  []T
	indexBuffer      []int
	minPriority      int
	sumProbability   float64
	entropy          float64
	version          int
}

func NewVariable[T comparable](name string, initialValues []DomainValue[T]) *Variable[T] {
	indices := make([]*index, len(initialValues))
	values := make([]T, len(initialValues))

	for idx, value := range initialValues {
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
		values[idx] = value.Value
	}

	variable := &Variable[T]{
		name:             name,
		indices:          indices,
		values:           values,
		indexBuffer:      make([]int, 0, len(indices)),
		availableValues:  make([]T, 0, len(indices)),
		availableIndices: make([]int, 0, len(indices)),
		version:          0,
	}

	variable.update()

	return variable
}

func (v *Variable[T]) GetName() string {
	return v.name
}

func (v *Variable[T]) AllowedValues() []T {
	return v.availableValues
}

func (v *Variable[T]) IsValueAllowed(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

func (v *Variable[T]) Exists(check func(a T) bool) bool {
	for _, availableValue := range v.availableValues {
		if check(availableValue) {
			return true
		}
	}
	return false
}

func (v *Variable[T]) ForEach(check func(a T) bool) bool {
	for _, availableValue := range v.availableValues {
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
		return v.values[v.availableIndices[0]]
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
	for _, availableIndex := range v.availableIndices {
		if v.values[availableIndex] == value {
			return v.Update(probabilityFactor, priority, availableIndex)
		}
	}
	return DoNothing
}

func (v *Variable[T]) AssignByValue(value T) Mutation {
	for _, availableIndex := range v.availableIndices {
		if v.values[availableIndex] == value {
			return v.Assign(availableIndex)
		}
	}
	return DoNothing
}

func (v *Variable[T]) ExcludeBy(shouldBan func(T) bool) Mutation {
	v.indexBuffer = v.indexBuffer[:0]
	for _, availableIndex := range v.availableIndices {
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

func NewVariableFromValues[T comparable](name string, values []T) *Variable[T] {
	return NewVariable[T](name, AsDomainValues(values...))
}

func (v *Variable[T]) IsUnassigned() bool {
	return len(v.availableIndices) > 1
}

func (v *Variable[T]) IsAssigned() bool {
	return len(v.availableIndices) == 1
}

func (v *Variable[T]) IsInContradiction() bool {
	return len(v.availableIndices) == 0
}

func (v *Variable[T]) IndexPriority(index int) int {
	return v.indices[index].priority
}

func (v *Variable[T]) IndexProbability(index int) float64 {
	return v.indices[index].probability
}

func (v *Variable[T]) Entropy() float64 {
	if !math.IsInf(v.entropy, +1) {
		return v.entropy
	}

	if v.sumProbability == 0.0 {
		v.entropy = math.Inf(-1)
		return v.entropy
	}

	entropy := 0.0
	for _, idx := range v.indices {
		if idx.isBanned || idx.priority != v.minPriority {
			continue
		}
		weightedProb := idx.probability / v.sumProbability
		entropy += weightedProb * math.Log2(weightedProb)
	}
	v.entropy = -entropy
	return v.entropy
}

func (v *Variable[T]) Assign(index int) Mutation {
	if index >= len(v.indices) {
		return v.Contradict()
	}

	v.indexBuffer = v.indexBuffer[:0]
	for _, availableIndex := range v.availableIndices {
		if availableIndex == index {
			continue
		}
		v.indexBuffer = append(v.indexBuffer, availableIndex)
	}

	return v.Exclude(v.indexBuffer...)
}

func (v *Variable[T]) Exclude(indices ...int) Mutation {
	return v.Update(0.0, 0, indices...)
}

func (v *Variable[T]) Contradict() Mutation {
	v.indexBuffer = v.indexBuffer[:0]
	for _, availableIndex := range v.availableIndices {
		v.indexBuffer = append(v.indexBuffer, availableIndex)
	}
	return v.Exclude(v.indexBuffer...)
}

func (v *Variable[T]) UpdatePriority(value int, indices ...int) Mutation {
	return v.Update(1.0, value, indices...)
}

func (v *Variable[T]) UpdateProbability(factor float64, indices ...int) Mutation {
	return v.Update(factor, 0, indices...)
}

func (v *Variable[T]) Update(probabilityFactory float64, priority int, indices ...int) Mutation {
	if len(indices) == 0 {
		return DoNothing
	}
	return Mutation{
		domain:      v,
		indices:     indices,
		probability: probabilityFactory,
		priority:    priority,
	}
}

func (v *Variable[T]) setId(id int) {
	v.id = id
}

func (v *Variable[T]) getId() int {
	return v.id
}

func (v *Variable[T]) numIndices() int {
	return len(v.indices)
}

func (v *Variable[T]) getMinPriority() int {
	return v.minPriority
}

func (v *Variable[T]) getIndex(i int) *index {
	return v.indices[i]
}

func (v *Variable[T]) setIndex(i int, idx *index) {
	v.indices[i] = idx
}

func (v *Variable[T]) getVersion() int {
	return v.version
}

func (v *Variable[T]) update() {
	v.version++

	v.sumProbability = 0.0
	v.minPriority = math.MaxInt
	v.entropy = math.Inf(+1)
	v.availableValues = v.availableValues[:0]
	v.availableIndices = v.availableIndices[:0]

	for i, idx := range v.indices {
		if !idx.isBanned {
			v.availableIndices = append(v.availableIndices, i)
			v.availableValues = append(v.availableValues, v.values[i])
		}
		if !idx.isBanned && idx.priority < v.minPriority {
			v.minPriority = idx.priority
		}
	}

	for _, idx := range v.indices {
		if !idx.isBanned && idx.priority == v.minPriority {
			v.sumProbability += idx.probability
		}
	}
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

func DomainsOf[T comparable](vars ...*Variable[T]) []Domain {
	domains := make([]Domain, 0, len(vars))
	for _, v := range vars {
		domains = append(domains, Domain(v))
	}
	return domains
}
