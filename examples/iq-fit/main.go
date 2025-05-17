package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
)

type Cell struct {
	X, Y  int
	Shape *propagator.Variable[int]
}

func main() {

	rows := 3
	columns := 3

	csp := propagator.NewProblem()

	var grid []*Cell

	for row := 0; row < rows; row++ {
		for col := 0; col < columns; col++ {
			cell := &Cell{
				X:     col,
				Y:     row,
				Shape: propagator.AddVariableFromValues(csp, fmt.Sprintf("%d,%d", col, row), []int{1, 2, 3}),
			}
			grid = append(grid, cell)
		}
	}

	model := csp.Model()
	fmt.Printf("model %v\n", model)

	solver := propagator.NewSolver()

	if !solver.Solve(model) {
		panic("No solution!")
	}
}

type Grid struct {
	cells []*Cell
}

func (q Grid) Scope() []propagator.DomainId {
	ids := make([]propagator.DomainId, len(q.cells))
	for _, cell := range q.cells {
		ids = append(ids, propagator.IdOf(cell.Shape))
	}
	return ids
}

func (q Grid) Propagate(m *propagator.Mutator) {
	// TODO: implement
	// probably extend propagate to get the currently set domain ID.
}
