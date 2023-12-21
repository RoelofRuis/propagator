package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
	"strings"
)

// Queen models the position of a queen.
// Assume one queen in each row. Which column does each one go in?
type Queen struct {
	Row    int
	Column *propagator.Variable[int]
}

func main() {
	csp := propagator.NewProblem()

	numQueens := 8
	//numQueens := 400

	rows := make([]int, numQueens)
	for i := 0; i < numQueens; i++ {
		rows[i] = i
	}

	queens := make([]Queen, numQueens)
	for i := 0; i < numQueens; i++ {
		queens[i] = Queen{
			Column: propagator.AddVariableFromValues(csp, fmt.Sprintf("queen_%d", i), rows),
			Row:    i,
		}
	}

	for i := 0; i < numQueens; i++ {
		for j := i + 1; j < numQueens; j++ {
			csp.AddConstraint(QueenExclusion{queens[i], queens[j]})
		}
	}

	model := csp.Model()

	solver := propagator.NewSolver()

	if !solver.Solve(model) {
		panic("no solution!")
	}

	for i := 0; i < numQueens; i++ {
		c := queens[i].Column.GetAssignedValue()
		fmt.Printf("%sQ %s\n", strings.Repeat(". ", c), strings.Repeat(". ", numQueens-1-c))
	}
}

type QueenExclusion struct {
	A Queen
	B Queen
}

func (q QueenExclusion) Scope() []propagator.DomainId {
	return propagator.IdsOf(q.A.Column, q.B.Column)
}

func (q QueenExclusion) Propagate(m *propagator.Mutator) {
	if q.A.Column.IsAssigned() {
		m.Add(q.B.Column.ExcludeByValue(q.A.Column.GetAssignedValue()))

		colA := q.A.Column.GetAssignedValue()
		m.Add(q.B.Column.ExcludeBy(func(colB int) bool {
			return absDiff(colB, colA) == absDiff(q.B.Row, q.A.Row)
		}))
	}
	if q.B.Column.IsAssigned() {
		m.Add(q.A.Column.ExcludeByValue(q.B.Column.GetAssignedValue()))

		colB := q.B.Column.GetAssignedValue()
		m.Add(q.A.Column.ExcludeBy(func(colA int) bool {
			return absDiff(colA, colB) == absDiff(q.A.Row, q.B.Row)
		}))
	}
}

func absDiff(x, y int) int {
	if x > y {
		return x - y
	} else {
		return y - x
	}
}
