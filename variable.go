package propagator

import (
	"math"
)

// TODO: try to move domain to model
type Domain struct {
	name string

	// id will be set by the model
	id               int // TODO: Or interface declaring 'something with domain id'
	indices          []*index
	availableIndices []int
	sumProbability   float64
	minPriority      int
	version          int
	indexBuffer      []int
}

func (d *Domain) update() {
	d.version++

	d.sumProbability = 0.0
	d.minPriority = math.MaxInt
	//d.entropy = math.Inf(+1) TODO: remove
	//d.availableValues = d.availableValues[:0] // TODO: see if we need these all the time or only when reading variable values and see how we can otherwise 'pass this update to variable'
	d.availableIndices = d.availableIndices[:0]

	for i, idx := range d.indices {
		if !idx.isBanned {
			d.availableIndices = append(d.availableIndices, i)
			//d.availableValues = append(d.availableValues, d.values[i]) // TODO: idem
		}
		if !idx.isBanned && idx.priority < d.minPriority {
			d.minPriority = idx.priority
		}
	}

	for _, idx := range d.indices {
		if !idx.isBanned && idx.priority == d.minPriority {
			d.sumProbability += idx.probability
		}
	}
}

func (d *Domain) getIndex(i int) *index {
	return d.indices[i]
}

func (d *Domain) setIndex(i int, idx *index) {
	d.indices[i] = idx
}

func (d *Domain) Assign(index int) Mutation {
	if index >= len(d.indices) {
		return d.Contradict()
	}

	d.indexBuffer = d.indexBuffer[:0]
	for _, availableIndex := range d.availableIndices {
		if availableIndex == index {
			continue
		}
		d.indexBuffer = append(d.indexBuffer, availableIndex)
	}

	return d.Exclude(d.indexBuffer...)
}

func (d *Domain) Exclude(indices ...int) Mutation {
	return d.Update(0.0, 0, indices...)
}

func (d *Domain) Contradict() Mutation {
	d.indexBuffer = d.indexBuffer[:0]
	for _, availableIndex := range d.availableIndices {
		d.indexBuffer = append(d.indexBuffer, availableIndex)
	}
	return d.Exclude(d.indexBuffer...)
}

func (d *Domain) Update(probabilityFactory float64, priority int, indices ...int) Mutation {
	if len(indices) == 0 {
		return DoNothing
	}
	return Mutation{
		domain:      d,
		indices:     indices,
		probability: probabilityFactory,
		priority:    priority,
	}
}

func (d *Domain) IsAssigned() bool {
	return len(d.availableIndices) == 1
}

func (d *Domain) IsUnassigned() bool {
	return len(d.availableIndices) > 1
}

func (d *Domain) IsInContradiction() bool {
	return len(d.availableIndices) == 0
}

type Variable[T comparable] struct {
	*Domain
	values          []T
	availableValues []T
}

func NewVariable[T comparable](name string, initialValues []DomainValue[T]) *Variable[T] {
	indices := make([]*index, len(initialValues))
	values := make([]T, len(initialValues))

	for idx, value := range initialValues {
		indices[idx] = indexFactorySingleton.create(value.Probability, value.Priority)
		values[idx] = value.Value
	}

	variable := &Variable[T]{
		Domain: &Domain{
			name:             name,
			indices:          indices,
			availableIndices: make([]int, 0, len(indices)),
			version:          0,
			indexBuffer:      make([]int, 0, len(indices)),
		},
		values:          values,
		availableValues: make([]T, 0, len(indices)),
	}

	variable.update()

	return variable
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

func (v *Variable[T]) IndexPriority(index int) int {
	return v.indices[index].priority
}

func (v *Variable[T]) IndexProbability(index int) float64 {
	return v.indices[index].probability
}

func (v *Variable[T]) UpdatePriority(value int, indices ...int) Mutation {
	return v.Update(1.0, value, indices...)
}

func (v *Variable[T]) UpdateProbability(factor float64, indices ...int) Mutation {
	return v.Update(factor, 0, indices...)
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

func DomainsOf[T comparable](vars ...*Variable[T]) []*Domain {
	domains := make([]*Domain, 0, len(vars))
	for _, v := range vars {
		domains = append(domains, v.Domain)
	}
	return domains
}
