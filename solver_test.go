package propagator

import (
	"math"
	"testing"
)

func TestSolver_FindAll(t *testing.T) {
	csp := NewProblem()
	varA := AddVariableFromValues(csp, "A", []int{1, 2, 3})
	varB := AddVariableFromValues(csp, "B", []int{1, 2, 3})

	csp.AddConstraint(largerThan{varA, varB})

	model := csp.Model()

	var solutions [][2]int

	solver := NewSolver(
		WithSeed(0),
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
	csp := NewProblem()

	varA := AddVariableFromValues(csp, "A", []int{1, 2, 3, 4})
	varB := AddVariableFromValues(csp, "B", []int{1, 2, 3, 4})

	csp.AddConstraint(largerThan{varA, varB})

	model := csp.Model()

	var solutions [][2]int

	solver := NewSolver(
		WithSeed(0),
		FindNSolutions(3),
		SelectDomainsByMinEntropy(),
		SelectIndicesProbabilistically(),
		On(SolutionFound, func() {
			solutions = append(solutions, [2]int{varA.GetAssignedValue(), varB.GetAssignedValue()})
		}),
	)

	solver.Solve(model)

	if len(solutions) != 3 || solutions[0] != [2]int{4, 1} || solutions[1] != [2]int{4, 3} || solutions[2] != [2]int{4, 2} {
		t.Fatalf("wrong or missing solutions: %v", solutions)
	}
}

func TestSolve(t *testing.T) {
	for i := 0; i < 100; i++ {
		csp := NewProblem()
		varA := AddVariableFromValues(csp, "A", []int{0, 1})
		vara := AddVariableFromValues(csp, "a", []int{0, 1})
		varb := AddVariableFromValues(csp, "b", []int{0, 1})
		varB := AddVariableFromValues(csp, "B", []int{0, 1})

		variables := []*Variable[int]{varA, vara, varb, varB}

		csp.AddConstraint(equals{varA, vara})
		csp.AddConstraint(equals{varB, varb})
		csp.AddConstraint(constraint{vara, varb})

		model := csp.Model()

		solver := NewSolver(
			WithSeed(int64(i)),
		)

		success := solver.Solve(model)

		if !success {
			t.Fatalf("Failed to solve [RUN=%d]", i)
		}

		for _, v := range variables {
			if !v.IsAssigned() {
				t.Fatalf("Failed to fix %s [RUN=%d]", v.Domain.Name(), i)
			} else if !(v.GetAssignedValue() == 1) {
				t.Fatalf("Invalid value for %s [RUN=%d]", v.Domain.Name(), i)
			}
		}
	}
}

func TestSolve_Hidden(t *testing.T) {
	csp := NewProblem()

	varA := AddVariableFromValues(csp, "A", []int{1, 2})
	varB := AddVariableFromValues(csp, "B", []int{1, 2})
	varC := AddHiddenVariableFromValues(csp, "C", []int{1, 2, 3, 4})

	csp.AddConstraint(largerThan{varB, varA})
	csp.AddConstraint(largerThan{varC, varB})

	model := csp.Model()

	solver := NewSolver()

	solved := solver.Solve(model)

	if !solved {
		t.Fatalf("Failed to find solution")
	}

	if !varA.IsAssigned() {
		t.Fatalf("Variable A should have a solution")
	}
	if varA.GetAssignedValue() != 1 {
		t.Fatalf("Variable A has wrong solution")
	}

	if !varB.IsAssigned() {
		t.Fatalf("Variable B should have a solution")
	}
	if varB.GetAssignedValue() != 2 {
		t.Fatalf("Variable B has wrong solution")
	}

	if varC.IsAssigned() {
		t.Fatalf("Variable C should not be solved because it is hidden")
	}
	if len(varC.AvailableValues()) != 2 || varC.AvailableValues()[0] != 3 || varC.AvailableValues()[1] != 4 {
		t.Fatalf("Variable C has wrong solution")
	}
}

type largerThan struct {
	a *Variable[int]
	b *Variable[int]
}

func (e largerThan) Scope() []DomainId {
	return IdsOf(e.a, e.b)
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
	for _, stateA := range e.a.AvailableValues() {
		if stateA <= minB {
			m.Add(e.a.ExcludeByValue(stateA))
		}
	}
	for _, stateB := range e.b.AvailableValues() {
		if stateB >= maxA {
			m.Add(e.b.ExcludeByValue(stateB))
		}
	}
}

type equals struct {
	a *Variable[int]
	b *Variable[int]
}

func (e equals) Scope() []DomainId {
	return IdsOf(e.a, e.b)
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

func (c constraint) Scope() []DomainId {
	return IdsOf(c.a, c.b)
}

func (c constraint) Propagate(m *Mutator) {
	if c.a.IsAssigned() && c.b.IsAssigned() {
		if !(c.a.GetAssignedValue() == 1 && c.b.GetAssignedValue() == 1) {
			m.Add(c.a.Contradict(), c.b.Contradict())
		}
	}
}
