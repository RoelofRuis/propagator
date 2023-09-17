package propagator

import (
	"math"
	"testing"
)

func TestSolver_FindAll(t *testing.T) {
	varA := NewVariableFromValues("A", []int{1, 2, 3})
	varB := NewVariableFromValues("B", []int{1, 2, 3})

	builder := BuildModel()
	builder.AddDomain(varA.Domain)
	builder.AddDomain(varB.Domain)

	builder.AddConstraint(largerThan{varA, varB})

	model := builder.Build()

	var solutions [][2]int

	solver := NewSolver(
		FindAllSolutions(),
		On(SolutionFound, func() {
			solutions = append(solutions, [2]int{varA.GetFixedValue(), varB.GetFixedValue()})
		}),
	)

	solver.Solve(model)

	if len(solutions) != 3 || solutions[0] != [2]int{3, 1} || solutions[1] != [2]int{3, 2} || solutions[2] != [2]int{2, 1} {
		t.Fatalf("wrong or missing solutions: %v", solutions)
	}
}

func TestSolver_FindFirstN(t *testing.T) {
	varA := NewVariableFromValues("A", []int{1, 2, 3, 4})
	varB := NewVariableFromValues("B", []int{1, 2, 3, 4})

	builder := BuildModel()
	builder.AddDomain(varA.Domain)
	builder.AddDomain(varB.Domain)

	builder.AddConstraint(largerThan{varA, varB})

	model := builder.Build()

	var solutions [][2]int

	solver := NewSolver(
		FindNSolutions(3),
		On(SolutionFound, func() {
			solutions = append(solutions, [2]int{varA.GetFixedValue(), varB.GetFixedValue()})
		}),
	)

	solver.Solve(model)

	if len(solutions) != 3 || solutions[0] != [2]int{4, 1} || solutions[1] != [2]int{4, 3} || solutions[2] != [2]int{4, 2} {
		t.Fatalf("wrong or missing solutions: %v", solutions)
	}
}

func TestSolver(t *testing.T) {
	for i := 0; i < 100; i++ {
		varA := NewVariableFromValues("A", []int{0, 1})
		vara := NewVariableFromValues("a", []int{0, 1})
		varb := NewVariableFromValues("b", []int{0, 1})
		varB := NewVariableFromValues("B", []int{0, 1})

		// vara and varb are hidden domains: they will not be picked/actively solved for.
		variables := []*Variable[int]{varA, vara, varb, varB}

		builder := BuildModel()
		builder.AddDomain(varA.Domain)
		builder.AddDomain(varB.Domain)

		builder.AddConstraint(equals{varA, vara})
		builder.AddConstraint(equals{varB, varb})
		builder.AddConstraint(constraint{vara, varb})

		model := builder.Build()

		solver := NewSolver(
			WithSeed(int64(i)),
		)

		success := solver.Solve(model)

		if !success {
			t.Fatalf("Failed to solve [RUN=%d]", i)
		}

		for _, v := range variables {
			if !v.IsFixed() {
				t.Fatalf("Failed to fix %s [RUN=%d]", v.Name, i)
			} else if !(v.GetFixedValue() == 1) {
				t.Fatalf("Invalid value for %s [RUN=%d]", v.Name, i)
			}
		}
	}
}

type largerThan struct {
	a *Variable[int]
	b *Variable[int]
}

func (e largerThan) Scope() []*Domain {
	return []*Domain{e.a.Domain, e.b.Domain}
}

func (e largerThan) Propagate(m *Mutator) {
	maxA := math.MinInt
	minB := math.MaxInt
	for _, valA := range e.a.AvailableValues() {
		if valA > maxA {
			maxA = valA
		}
	}
	for _, valB := range e.b.AvailableValues() {
		if valB < minB {
			minB = valB
		}
	}
	for _, stateA := range e.a.AvailableStates() {
		if stateA.Value <= minB {
			m.Add(e.a.Ban(stateA.Index))
		}
	}
	for _, stateB := range e.b.AvailableStates() {
		if stateB.Value >= maxA {
			m.Add(e.b.Ban(stateB.Index))
		}
	}
}

type equals struct {
	a *Variable[int]
	b *Variable[int]
}

func (e equals) Scope() []*Domain {
	return []*Domain{e.a.Domain, e.b.Domain}
}

func (e equals) Propagate(m *Mutator) {
	if e.a.IsFixed() {
		m.Add(e.b.FixByValue(e.a.GetFixedValue()))
	}
	if e.b.IsFixed() {
		m.Add(e.a.FixByValue(e.b.GetFixedValue()))
	}
}

type constraint struct {
	a *Variable[int]
	b *Variable[int]
}

func (c constraint) Scope() []*Domain {
	return []*Domain{c.a.Domain, c.b.Domain}
}

func (c constraint) Propagate(m *Mutator) {
	if c.a.IsFixed() && c.b.IsFixed() {
		if !(c.a.GetFixedValue() == 1 && c.b.GetFixedValue() == 1) {
			m.Add(c.a.Contradict(), c.b.Contradict())
		}
	}
}
