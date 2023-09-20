package propagator

import "math"

type Domain2 interface {
	getId() int
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

type Variable2[T comparable] struct {
	id               int
	name             string
	indices          []*index
	availableIndices []int
	values           []T
	availableValues  []T
	indexBuffer      []int
	minPriority      int
	sumProbability   float64
	entropy          float64
	version          int

	// TODO: map[T]int misschien super handig?
}

func NewVariable2[T comparable](name string, initialValues []DomainValue[T]) *Variable2[T] {
	indices := make([]*index, len(initialValues))
	values := make([]T, len(initialValues))

	for idx, value := range initialValues {
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
		values[idx] = value.Value
	}

	variable := &Variable2[T]{
		name:             name,
		indices:          indices,
		availableIndices: make([]int, 0, len(indices)),
		values:           values,
		availableValues:  make([]T, 0, len(indices)),
		indexBuffer:      make([]int, 0, len(indices)),
		version:          0,
	}

	variable.update()

	return variable
}

func (v Variable2[T]) GetName() string {
	return v.name
}

func (v Variable2[T]) AllowedValues() []T {
	return v.availableValues
}

func (v Variable2[T]) IsValueAllowed(value T) bool {
	return v.Exists(func(a T) bool { return a == value })
}

func (v Variable2[T]) Exists(check func(a T) bool) bool {
	for _, availableValue := range v.availableValues {
		if check(availableValue) {
			return true
		}
	}
	return false
}

func (v Variable2[T]) ForEach(check func(a T) bool) bool {
	for _, availableValue := range v.availableValues {
		if !check(availableValue) {
			return false
		}
	}
	return true
}

func (v Variable2[T]) HasAnyOf(values ...T) bool {
	return v.Exists(func(a T) bool {
		for _, value := range values {
			if a == value {
				return true
			}
		}
		return false
	})
}

func (v Variable2[T]) GetAssignedValue() T {
	if !v.IsAssigned() {
		panic("Trying to GetAssignedValue on non fixed variable. Use IsAssigned to check.")
	}
	return v.availableValues[0]
}

func (v Variable2[T]) UpdatePriorityByValue(priority int, value T) Mutation {
	return v.UpdateByValue(1.0, priority, value)
}

func (v Variable2[T]) UpdateProbabilityByValue(factory float64, value T) Mutation {
	return v.UpdateByValue(factory, 0, value)
}

func (v Variable2[T]) UpdateByValue(probabilityFactor float64, priority int, value T) Mutation {
	for _, availableIndex := range v.availableIndices {
		if v.values[availableIndex] == value {
			return v.Update(probabilityFactor, priority, availableIndex)
		}
	}
	return DoNothing
}

func (v Variable2[T]) AssignByValue(value T) Mutation {
	for _, availableIndex := range v.availableIndices {
		if v.values[availableIndex] == value {
			return v.Assign(availableIndex)
		}
	}
	return DoNothing
}

func (v Variable2[T]) ExcludeBy(shouldBan func(T) bool) Mutation {
	v.indexBuffer = v.indexBuffer[:0]
	for _, availableIndex := range v.availableIndices {
		if shouldBan(v.values[availableIndex]) {
			v.indexBuffer = append(v.indexBuffer, availableIndex)
		}
	}
	return v.Exclude(v.indexBuffer...)
}

func (v Variable2[T]) ExcludeByValue(values ...T) Mutation {
	return v.ExcludeBy(func(a T) bool {
		for _, value := range values {
			if value == a {
				return true
			}
		}
		return false
	})
}

func NewVariable2FromValues[T comparable](name string, values []T) *Variable2[T] {
	return NewVariable2[T](name, AsDomainValues(values...))
}

func (v Variable2[T]) IsUnassigned() bool {
	return len(v.availableIndices) > 1
}

func (v Variable2[T]) IsAssigned() bool {
	return len(v.availableIndices) == 1
}

func (v Variable2[T]) IsInContradiction() bool {
	return len(v.availableIndices) == 0
}

func (v Variable2[T]) IndexPriority(index int) int {
	return v.indices[index].priority
}

func (v Variable2[T]) IndexProbability(index int) float64 {
	return v.indices[index].probability
}

func (v Variable2[T]) GetAssignedIndex() int { // Is this required?
	if v.IsAssigned() {
		return -1
	}
	return v.availableIndices[0]
}

func (v Variable2[T]) Entropy() float64 {
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

func (v Variable2[T]) Assign(index int) Mutation {
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

func (v Variable2[T]) Exclude(indices ...int) Mutation {
	return v.Update(0.0, 0, indices...)
}

func (v Variable2[T]) Contradict() Mutation {
	return v.Exclude(v.availableIndices...)
}

func (v Variable2[T]) UpdatePriority(value int, indices ...int) Mutation {
	return v.Update(1.0, value, indices...)
}

func (v Variable2[T]) UpdateProbability(factor float64, indices ...int) Mutation {
	return v.Update(factor, 0, indices...)
}

func (v Variable2[T]) Update(probabilityFactory float64, priority int, indices ...int) Mutation {
	if len(indices) == 0 {
		return DoNothing
	}
	return Mutation{
		// TODO: FIX domain:      v,
		indices:     indices,
		probability: probabilityFactory,
		priority:    priority,
	}
}

func (v Variable2[T]) getId() int {
	return v.id
}

func (v Variable2[T]) numIndices() int {
	return len(v.indices)
}

func (v Variable2[T]) getMinPriority() int {
	return v.minPriority
}

func (v Variable2[T]) getIndex(i int) *index {
	return v.indices[i]
}

func (v Variable2[T]) setIndex(i int, idx *index) {
	v.indices[i] = idx
}

func (v Variable2[T]) getVersion() int {
	return v.version
}

func (v Variable2[T]) update() {
	v.version++

	v.sumProbability = 0.0
	v.minPriority = math.MaxInt
	v.entropy = math.Inf(+1)

	for i, idx := range v.indices {
		if !idx.isBanned {
			v.availableIndices = append(v.availableIndices, i)
		}
		if !idx.isBanned && idx.priority < v.minPriority {
			v.minPriority = idx.priority
		}
	}

	v.availableIndices = v.availableIndices[:0]
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