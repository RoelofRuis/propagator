package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
	"os"
	"strconv"
	"strings"
)

type Cell = propagator.Variable[int]

func main() {
	content, err := os.ReadFile("examples/calcudoku/puzzle3.txt")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	size, _ := strconv.ParseInt(lines[0], 10, 64)

	gridSize := int(size)

	values := make([]int, gridSize)
	cells := make([][]*Cell, gridSize)
	rows := make([]House, gridSize)
	cols := make([]House, gridSize)
	for i := 0; i < gridSize; i++ {
		values[i] = i + 1
		cells[i] = make([]*Cell, gridSize)
		rows[i] = House{}
		cols[i] = House{}
	}

	csp := propagator.NewProblem()

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			cell := propagator.AddVariableFromValues(csp, fmt.Sprintf("%d,%d", x, y), values)
			cells[x][y] = cell
			rows[y].Cells = append(rows[y].Cells, cell)
			cols[x].Cells = append(cols[x].Cells, cell)
		}
	}

	for i := 0; i < gridSize; i++ {
		csp.AddConstraint(rows[i])
		csp.AddConstraint(cols[i])
	}

	for i := 1; i < len(lines); i++ {
		parts := strings.Split(lines[i], " ")
		opp := parts[0]
		value, _ := strconv.ParseInt(parts[1], 10, 64)
		var cageCells []*Cell
		for _, coordPair := range parts[2:] {
			xy := strings.Split(coordPair, ".")
			x, _ := strconv.ParseInt(xy[0], 10, 64)
			y, _ := strconv.ParseInt(xy[1], 10, 64)
			cageCells = append(cageCells, cells[x][y])
		}
		switch opp {
		case ".":
			csp.AddConstraint(FixedCage{Cell: cageCells[0], Value: int(value)})
			break
		case "+":
			csp.AddConstraint(SumCage{Cells: cageCells, Value: int(value)})
			break
		case "*":
			csp.AddConstraint(ProdCage{Cells: cageCells, Value: int(value)})
			break
		case "-":
			csp.AddConstraint(SubCage{Cells: cageCells, Value: int(value)})
			break
		case "/":
			csp.AddConstraint(DivCage{Cells: cageCells, Value: int(value)})
			break
		}
	}

	model := csp.Model()

	solver := propagator.NewSolver(
		// FIXME: with this problem the solution time seems to depend very much on the seed.
		propagator.WithSeed(0),
		propagator.LogInfo(),
	)

	if !solver.Solve(model) {
		panic("No solution!")
	}

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			fmt.Printf("[%d]", cells[x][y].GetAssignedValue())
		}
		fmt.Printf("\n")
	}
}
