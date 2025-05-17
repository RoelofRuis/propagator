package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
)

type Placement struct {
	X           int
	Y           int
	Orientation int
}

type Shape struct {
	Placement *propagator.Variable[Placement]
}

func main() {
	rows := 3
	columns := 3
	// rows := 5
	// columns := 10

	var domain []Placement
	for row := 0; row < rows; row++ {
		for col := 0; col < columns; col++ {
			for orientation := 0; orientation < 3; orientation++ {
				domain = append(domain, Placement{X: col, Y: row, Orientation: orientation})
			}
		}
	}

	csp := propagator.NewProblem()

	var shapes []Shape
	for s := 0; s < 3; s++ {
		shapes = append(shapes, Shape{
			Placement: propagator.AddVariableFromValues(csp, fmt.Sprintf("shape %d", s), domain),
		})
	}

	csp.AddConstraint(Grid{shapes})

	model := csp.Model()

	solver := propagator.NewSolver()

	if !solver.Solve(model) {
		panic("No solution!")
	}

	for i, shape := range shapes {
		c := shape.Placement.GetAssignedValue()
		fmt.Printf("%d: %v\n", i, c)
	}
}

type Grid struct {
	shapes []Shape
}

func (q Grid) Scope() []propagator.DomainId {
	ids := make([]propagator.DomainId, len(q.shapes))
	for _, shape := range q.shapes {
		ids = append(ids, propagator.IdOf(shape.Placement))
	}
	return ids
}

func (q Grid) Propagate(m *propagator.Mutator) {
	// TODO: implement

}
