package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Cell = propagator.Variable[int]

type AdjacentCells struct {
	CellA *Cell
	CellB *Cell
}

func (a AdjacentCells) Scope() []propagator.DomainId {
	return propagator.IdsOf(a.CellA, a.CellB)
}

func (a AdjacentCells) Propagate(mutator *propagator.Mutator) {
	if a.CellA.IsAssigned() {
		mutator.Add(a.CellB.ExcludeByValue(a.CellA.GetAssignedValue()))
	} else if a.CellB.IsAssigned() {
		mutator.Add(a.CellA.ExcludeByValue(a.CellB.GetAssignedValue()))
	}
}

type Block struct {
	Size  int
	Cells []*Cell
}

func (b Block) Scope() []propagator.DomainId {
	return propagator.IdsOf(b.Cells...)
}

func (b Block) Propagate(mutator *propagator.Mutator) {
	for _, i := range b.Cells {
		if !i.IsAssigned() {
			continue
		}
		for _, j := range b.Cells {
			if j == i {
				continue
			}
			mutator.Add(j.ExcludeByValue(i.GetAssignedValue()))
		}
	}
}

type CellData struct {
	value   int
	blockId int
}

func main() {
	content, err := os.ReadFile("examples/tectonic/puzzle3.txt")
	if err != nil {
		log.Panic(err)
	}

	lines := strings.Split(string(content), "\n")
	initialCells := make([][]CellData, 0, len(lines))
	blocks := make(map[int]Block)

	for _, line := range lines {
		var initialLine []CellData
		cells := strings.Split(strings.Join(strings.Fields(line), " "), " ")
		for _, cell := range cells {
			parts := strings.Split(cell, ":")
			value, err := strconv.Atoi(parts[0])
			if err != nil {
				value = 0
			}
			blockId, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Panic(err)
			}
			block, has := blocks[blockId]
			if !has {
				block = Block{}
			}
			block.Size++
			blocks[blockId] = block
			initialLine = append(initialLine, CellData{value: value, blockId: blockId})
		}
		initialCells = append(initialCells, initialLine)
	}

	csp := propagator.NewProblem()

	var field [][]*Cell

	for x, cellLine := range initialCells {
		var fieldLine []*Cell
		for y, cellData := range cellLine {
			var cell *Cell
			block := blocks[cellData.blockId]
			if cellData.value != 0 {
				cell = propagator.AddVariableFromValues(csp, fmt.Sprintf("%d,%d", x, y), []int{cellData.value})
			} else {
				var values []int
				for i := 1; i <= block.Size; i++ {
					values = append(values, i)
				}
				cell = propagator.AddVariableFromValues(csp, fmt.Sprintf("%d", x), values)
			}
			block.Cells = append(block.Cells, cell)
			blocks[cellData.blockId] = block
			fieldLine = append(fieldLine, cell)
		}
		field = append(field, fieldLine)
	}

	for _, block := range blocks {
		csp.AddConstraint(block)
	}

	for y, fieldLine := range field {
		for x, cell := range fieldLine {
			var xCell, yCell bool
			if (x + 1) < len(fieldLine) {
				adjacentCell := field[y][x+1]
				csp.AddConstraint(AdjacentCells{cell, adjacentCell})
				xCell = true
			}
			if (y + 1) < len(field) {
				adjacentCell := field[y+1][x]
				csp.AddConstraint(AdjacentCells{cell, adjacentCell})
				yCell = true
			}
			if xCell && yCell {
				adjacentCell := field[y+1][x+1]
				csp.AddConstraint(AdjacentCells{cell, adjacentCell})
			}
			if (x-1) > 0 && yCell {
				adjacentCell := field[y+1][x-1]
				csp.AddConstraint(AdjacentCells{cell, adjacentCell})
			}
		}
	}

	model := csp.Model()

	solver := propagator.NewSolver()

	start := time.Now()
	if !solver.Solve(model) {
		panic("No solution!")
	}

	fmt.Printf("Solution found in %s:\n", time.Since(start))
	for _, fieldLine := range field {
		for _, cell := range fieldLine {
			fmt.Printf("[%d]", cell.GetAssignedValue())
		}
		fmt.Printf("\n")
	}
}
