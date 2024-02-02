package propagator

// ProbabilityModifiers stores probability modifiers together with a running product.
type ProbabilityModifiers struct {
	Data               map[constraintId]Probability
	ProbabilityProduct Probability
}

func NewProbabilityModifiers() *ProbabilityModifiers {
	return &ProbabilityModifiers{
		Data:               make(map[constraintId]Probability),
		ProbabilityProduct: 1.0,
	}
}

func (m *ProbabilityModifiers) Clone() *ProbabilityModifiers {
	clone := make(map[constraintId]Probability)
	for k, v := range m.Data {
		clone[k] = v
	}
	return &ProbabilityModifiers{Data: clone, ProbabilityProduct: m.ProbabilityProduct}
}

func (m *ProbabilityModifiers) Insert(k constraintId, p Probability) {
	oldValue, exists := m.Data[k]
	if exists {
		m.ProbabilityProduct = m.ProbabilityProduct / oldValue
	}
	m.Data[k] = p
	m.ProbabilityProduct = m.ProbabilityProduct * p
}

// PriorityModifiers stores priority modifiers together with a running sum.
type PriorityModifiers struct {
	Data        map[constraintId]Priority
	PrioritySum Priority
}

func NewPriorityModifiers() *PriorityModifiers {
	return &PriorityModifiers{
		Data:        make(map[constraintId]Priority),
		PrioritySum: 0,
	}
}

func (m *PriorityModifiers) Clone() *PriorityModifiers {
	clone := make(map[constraintId]Priority)
	for k, v := range m.Data {
		clone[k] = v
	}
	return &PriorityModifiers{Data: clone, PrioritySum: m.PrioritySum}
}

func (m *PriorityModifiers) Insert(k constraintId, p Priority) {
	oldValue, exists := m.Data[k]
	if exists {
		m.PrioritySum = m.PrioritySum - oldValue
	}
	m.Data[k] = p
	m.PrioritySum = m.PrioritySum + p
}
