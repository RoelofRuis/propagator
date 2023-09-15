package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
)

func main() {
	letters := []string{"S", "E", "N", "D", "M", "O", "R", "Y"}

	digits := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	builder := propagator.BuildModel()
	variables := make(map[string]*propagator.Variable[int])
	allDifferent := AllDifferent{}
	for _, letter := range letters {
		variable := propagator.NewVariableFromValues[int](letter, digits)
		builder.AddDomain(variable.Domain)
		variables[letter] = variable
		allDifferent.Variables = append(allDifferent.Variables, variable)
	}

	builder.AddConstraint(allDifferent)

	n1 := Number{[]*propagator.Variable[int]{variables["S"], variables["E"], variables["N"], variables["D"]}}
	n2 := Number{[]*propagator.Variable[int]{variables["M"], variables["O"], variables["R"], variables["E"]}}
	n3 := Number{[]*propagator.Variable[int]{variables["M"], variables["O"], variables["N"], variables["E"], variables["Y"]}}

	builder.AddConstraint(n1)
	builder.AddConstraint(n2)
	builder.AddConstraint(n3)

	builder.AddConstraint(Sum{n1, n2, n3})

	model := builder.Build()

	solver := propagator.NewSolver(model)

	if !solver.Solve() {
		panic("no solution")
	}

	for _, letter := range letters {
		variable := variables[letter]
		fmt.Printf("%s: %d\n", letter, variable.GetFixedValue())
	}

	fmt.Printf("\n%d\n%d\n---- +\n%d", n1.Decimal(), n2.Decimal(), n3.Decimal())
}