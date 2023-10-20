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
	content, err := os.ReadFile("examples/sudoku/puzzle1.txt")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")

	gridSize := 9
	cells := make([][]*Cell, gridSize)
	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < gridSize; i++ {
		cells[i] = make([]*Cell, gridSize)
	}

	builder := propagator.BuildModel()

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			char := string(lines[y][x])
			var cell *Cell
			switch char {
			case ".":
				cell = propagator.NewVariableFromValues(fmt.Sprintf("%d,%d", x, y), values)
				break
			default:
				val, _ := strconv.ParseInt(char, 10, 64)
				cell = propagator.NewVariableFromValues(fmt.Sprintf("%d,%d", x, y), []int{int(val)})
			}
			cells[x][y] = cell
			builder.AddDomain(cell)
		}
	}

	builder.AddConstraint(House{[]*Cell{cells[0][0], cells[1][0], cells[2][0], cells[3][0], cells[4][0], cells[5][0], cells[6][0], cells[7][0], cells[8][0]}})
	builder.AddConstraint(House{[]*Cell{cells[0][1], cells[1][1], cells[2][1], cells[3][1], cells[4][1], cells[5][1], cells[6][1], cells[7][1], cells[8][1]}})
	builder.AddConstraint(House{[]*Cell{cells[0][2], cells[1][2], cells[2][2], cells[3][2], cells[4][2], cells[5][2], cells[6][2], cells[7][2], cells[8][2]}})
	builder.AddConstraint(House{[]*Cell{cells[0][3], cells[1][3], cells[2][3], cells[3][3], cells[4][3], cells[5][3], cells[6][3], cells[7][3], cells[8][3]}})
	builder.AddConstraint(House{[]*Cell{cells[0][4], cells[1][4], cells[2][4], cells[3][4], cells[4][4], cells[5][4], cells[6][4], cells[7][4], cells[8][4]}})
	builder.AddConstraint(House{[]*Cell{cells[0][5], cells[1][5], cells[2][5], cells[3][5], cells[4][5], cells[5][5], cells[6][5], cells[7][5], cells[8][5]}})
	builder.AddConstraint(House{[]*Cell{cells[0][6], cells[1][6], cells[2][6], cells[3][6], cells[4][6], cells[5][6], cells[6][6], cells[7][6], cells[8][6]}})
	builder.AddConstraint(House{[]*Cell{cells[0][7], cells[1][7], cells[2][7], cells[3][7], cells[4][7], cells[5][7], cells[6][7], cells[7][7], cells[8][7]}})
	builder.AddConstraint(House{[]*Cell{cells[0][8], cells[1][8], cells[2][8], cells[3][8], cells[4][8], cells[5][8], cells[6][8], cells[7][8], cells[8][8]}})

	builder.AddConstraint(House{[]*Cell{cells[0][0], cells[0][1], cells[0][2], cells[0][3], cells[0][4], cells[0][5], cells[0][6], cells[0][7], cells[0][8]}})
	builder.AddConstraint(House{[]*Cell{cells[1][0], cells[1][1], cells[1][2], cells[1][3], cells[1][4], cells[1][5], cells[1][6], cells[1][7], cells[1][8]}})
	builder.AddConstraint(House{[]*Cell{cells[2][0], cells[2][1], cells[2][2], cells[2][3], cells[2][4], cells[2][5], cells[2][6], cells[2][7], cells[2][8]}})
	builder.AddConstraint(House{[]*Cell{cells[3][0], cells[3][1], cells[3][2], cells[3][3], cells[3][4], cells[3][5], cells[3][6], cells[3][7], cells[3][8]}})
	builder.AddConstraint(House{[]*Cell{cells[4][0], cells[4][1], cells[4][2], cells[4][3], cells[4][4], cells[4][5], cells[4][6], cells[4][7], cells[4][8]}})
	builder.AddConstraint(House{[]*Cell{cells[5][0], cells[5][1], cells[5][2], cells[5][3], cells[5][4], cells[5][5], cells[5][6], cells[5][7], cells[5][8]}})
	builder.AddConstraint(House{[]*Cell{cells[6][0], cells[6][1], cells[6][2], cells[6][3], cells[6][4], cells[6][5], cells[6][6], cells[6][7], cells[6][8]}})
	builder.AddConstraint(House{[]*Cell{cells[7][0], cells[7][1], cells[7][2], cells[7][3], cells[7][4], cells[7][5], cells[7][6], cells[7][7], cells[7][8]}})
	builder.AddConstraint(House{[]*Cell{cells[8][0], cells[8][1], cells[8][2], cells[8][3], cells[8][4], cells[8][5], cells[8][6], cells[8][7], cells[8][8]}})

	builder.AddConstraint(House{[]*Cell{cells[0][0], cells[1][0], cells[2][0], cells[0][1], cells[1][1], cells[2][1], cells[0][2], cells[1][2], cells[2][2]}})
	builder.AddConstraint(House{[]*Cell{cells[3][0], cells[4][0], cells[5][0], cells[3][1], cells[4][1], cells[5][1], cells[3][2], cells[4][2], cells[5][2]}})
	builder.AddConstraint(House{[]*Cell{cells[6][0], cells[7][0], cells[8][0], cells[6][1], cells[7][1], cells[8][1], cells[6][2], cells[7][2], cells[8][2]}})
	builder.AddConstraint(House{[]*Cell{cells[0][3], cells[1][3], cells[2][3], cells[0][4], cells[1][4], cells[2][4], cells[0][5], cells[1][5], cells[2][5]}})
	builder.AddConstraint(House{[]*Cell{cells[3][3], cells[4][3], cells[5][3], cells[3][4], cells[4][4], cells[5][4], cells[3][5], cells[4][5], cells[5][5]}})
	builder.AddConstraint(House{[]*Cell{cells[6][3], cells[7][3], cells[8][3], cells[6][4], cells[7][4], cells[8][4], cells[6][5], cells[7][5], cells[8][5]}})
	builder.AddConstraint(House{[]*Cell{cells[0][6], cells[1][6], cells[2][6], cells[0][7], cells[1][7], cells[2][7], cells[0][8], cells[1][8], cells[2][8]}})
	builder.AddConstraint(House{[]*Cell{cells[3][6], cells[4][6], cells[5][6], cells[3][7], cells[4][7], cells[5][7], cells[3][8], cells[4][8], cells[5][8]}})
	builder.AddConstraint(House{[]*Cell{cells[6][6], cells[7][6], cells[8][6], cells[6][7], cells[7][7], cells[8][7], cells[6][8], cells[7][8], cells[8][8]}})

	model := builder.Build()

	solver := propagator.NewSolver()

	if !solver.Solve(model) {
		panic("no solution!")
	}

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			fmt.Printf("[%d]", cells[x][y].GetAssignedValue())
		}
		fmt.Printf("\n")
	}
}
