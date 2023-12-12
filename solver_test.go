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
			solutions = append(solutions, [2]int{varA.GetAssignedValue(), varB.GetAssignedValue()})
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
			solutions = append(solutions, [2]int{varA.GetAssignedValue(), varB.GetAssignedValue()})
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
			if !v.IsAssigned() {
				t.Fatalf("Failed to fix %s [RUN=%d]", model.NameOf(v.Domain), i)
			} else if !(v.GetAssignedValue() == 1) {
				t.Fatalf("Invalid value for %s [RUN=%d]", model.NameOf(v.Domain), i)
			}
		}
	}
}

type largerThan struct {
	a *Variable[int]
	b *Variable[int]
}

func (e largerThan) Scope() []*Domain {
	return DomainsOf(e.a, e.b)
}

func (e largerThan) Propagate(m *Mutator) {
	maxA := math.MinInt
	minB := math.MaxInt
	for _, valA := range e.a.AllowedValues() {
		if valA > maxA {
			maxA = valA
		}
	}
	for _, valB := range e.b.AllowedValues() {
		if valB < minB {
			minB = valB
		}
	}
	for _, stateA := range e.a.AllowedValues() {
		if stateA <= minB {
			m.Add(e.a.ExcludeByValue(stateA))
		}
	}
	for _, stateB := range e.b.AllowedValues() {
		if stateB >= maxA {
			m.Add(e.b.ExcludeByValue(stateB))
		}
	}
}

type equals struct {
	a *Variable[int]
	b *Variable[int]
}

func (e equals) Scope() []*Domain {
	return DomainsOf(e.a, e.b)
}

func (e equals) Propagate(m *Mutator) {
	if e.a.IsAssigned() {
		m.Add(e.b.AssignByValue(e.a.GetAssignedValue()))
	}
	if e.b.IsAssigned() {
		m.Add(e.a.AssignByValue(e.b.GetAssignedValue()))
	}
}

type constraint struct {
	a *Variable[int]
	b *Variable[int]
}

func (c constraint) Scope() []*Domain {
	return DomainsOf(c.a, c.b)
}

func (c constraint) Propagate(m *Mutator) {
	if c.a.IsAssigned() && c.b.IsAssigned() {
		if !(c.a.GetAssignedValue() == 1 && c.b.GetAssignedValue() == 1) {
			m.Add(c.a.Contradict(), c.b.Contradict())
		}
	}
}
