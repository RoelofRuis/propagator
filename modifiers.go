package propagator

// probabilityModifiers stores probability modifiers together with a running product.
type probabilityModifiers struct {
	Data               map[constraintId]Probability
	ProbabilityProduct Probability
}

func newProbabilityModifiers() *probabilityModifiers {
	return &probabilityModifiers{
		Data:               make(map[constraintId]Probability),
		ProbabilityProduct: 1.0,
	}
}

var oneProbabilityModifier = &probabilityModifiers{Data: nil, ProbabilityProduct: 1}
var zeroProbabilityModifier = &probabilityModifiers{Data: nil, ProbabilityProduct: 0}

func insertProbability(m *probabilityModifiers, c constraintId, p Probability) *probabilityModifiers {
	if p < 10e-10 {
		return zeroProbabilityModifier
	}
	if len(m.Data) == 0 && c == -1 && p == 1 {
		return oneProbabilityModifier
	}
	clone := make(map[constraintId]Probability, len(m.Data))
	for k, v := range m.Data {
		clone[k] = v
	}
	newModifiers := &probabilityModifiers{ProbabilityProduct: m.ProbabilityProduct}
	oldValue, exists := m.Data[c]
	if exists {
		newModifiers.ProbabilityProduct = newModifiers.ProbabilityProduct / oldValue
	}
	clone[c] = p
	newModifiers.ProbabilityProduct = newModifiers.ProbabilityProduct * p
	newModifiers.Data = clone
	return newModifiers
}

// priorityModifiers stores priority modifiers together with a running sum.
type priorityModifiers struct {
	Data        map[constraintId]Priority
	PrioritySum Priority
}

var zeroPriorityModifier = &priorityModifiers{Data: nil, PrioritySum: 0}

func newPriorityModifiers() *priorityModifiers {
	return &priorityModifiers{
		Data:        make(map[constraintId]Priority),
		PrioritySum: 0,
	}
}

func insertPriority(m *priorityModifiers, c constraintId, p Priority) *priorityModifiers {
	if len(m.Data) == 0 && c == -1 && p == 0 {
		return zeroPriorityModifier
	}
	clone := make(map[constraintId]Priority)
	for k, v := range m.Data {
		clone[k] = v
	}
	newModifiers := &priorityModifiers{PrioritySum: m.PrioritySum}
	oldValue, exists := m.Data[c]
	if exists {
		newModifiers.PrioritySum = newModifiers.PrioritySum - oldValue
	}
	newModifiers.PrioritySum = newModifiers.PrioritySum + p
	newModifiers.Data = clone
	return newModifiers
}

func (m *priorityModifiers) Insert(k constraintId, p Priority) {
	oldValue, exists := m.Data[k]
	if exists {
		m.PrioritySum = m.PrioritySum - oldValue
	}
	m.Data[k] = p
	m.PrioritySum = m.PrioritySum + p
}
