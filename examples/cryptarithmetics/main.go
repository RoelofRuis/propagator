package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
)

func main() {
	letters := []string{"S", "E", "N", "D", "M", "O", "R", "Y"}

	digits := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	csp := propagator.NewCSP()
	variables := make(map[string]*propagator.Variable[int])
	allDifferent := AllDifferent{}
	for _, letter := range letters {
		variable := propagator.AddVariableFromValues[int](csp, letter, digits)
		variables[letter] = variable
		allDifferent.Variables = append(allDifferent.Variables, variable)
	}

	csp.AddConstraint(allDifferent)

	n1 := Number{[]*propagator.Variable[int]{variables["S"], variables["E"], variables["N"], variables["D"]}}
	n2 := Number{[]*propagator.Variable[int]{variables["M"], variables["O"], variables["R"], variables["E"]}}
	n3 := Number{[]*propagator.Variable[int]{variables["M"], variables["O"], variables["N"], variables["E"], variables["Y"]}}

	csp.AddConstraint(n1)
	csp.AddConstraint(n2)
	csp.AddConstraint(n3)

	csp.AddConstraint(Sum{n1, n2, n3})

	model := csp.GetModel()

	solver := propagator.NewSolver()

	if !solver.Solve(model) {
		panic("no solution")
	}

	for _, letter := range letters {
		variable := variables[letter]
		fmt.Printf("%s: %d\n", letter, variable.GetAssignedValue())
	}

	fmt.Printf("\n%d\n%d\n---- +\n%d", n1.Decimal(), n2.Decimal(), n3.Decimal())
}
