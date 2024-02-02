package propagator

// ProbabilityModifiers stores probability modifiers together with a running product.
type ProbabilityModifiers struct {
	inner   map[constraintId]Probability
	product Probability
}

func NewProbabilityModifiers() *ProbabilityModifiers {
	return &ProbabilityModifiers{
		inner:   make(map[constraintId]Probability),
		product: 1.0,
	}
}

func (m *ProbabilityModifiers) Clone() *ProbabilityModifiers {
	clone := make(map[constraintId]Probability)
	for k, v := range m.inner {
		clone[k] = v
	}
	return &ProbabilityModifiers{inner: clone, product: m.product}
}

func (m *ProbabilityModifiers) Insert(k constraintId, p Probability) {
	oldValue, exists := m.inner[k]
	if exists {
		m.product = m.product / oldValue
	}
	m.inner[k] = p
	m.product = m.product * p
}

func (m *ProbabilityModifiers) Get(k constraintId) (Probability, bool) {
	value, exists := m.inner[k]
	return value, exists
}

func (m *ProbabilityModifiers) Product() Probability {
	return m.product
}

// PriorityModifiers stores priority modifiers together with a running sum.
type PriorityModifiers struct {
	inner map[constraintId]Priority
	sum   Priority
}

func NewPriorityModifiers() *PriorityModifiers {
	return &PriorityModifiers{
		inner: make(map[constraintId]Priority),
		sum:   0,
	}
}

func (m *PriorityModifiers) Clone() *PriorityModifiers {
	clone := make(map[constraintId]Priority)
	for k, v := range m.inner {
		clone[k] = v
	}
	return &PriorityModifiers{inner: clone, sum: m.sum}
}

func (m *PriorityModifiers) Insert(k constraintId, p Priority) {
	oldValue, exists := m.inner[k]
	if exists {
		m.sum = m.sum - oldValue
	}
	m.inner[k] = p
	m.sum = m.sum + p
}

func (m *PriorityModifiers) Get(k constraintId) (Priority, bool) {
	value, exists := m.inner[k]
	return value, exists
}

func (m *PriorityModifiers) Sum() Priority {
	return m.sum
}
