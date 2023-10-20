package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"
)

func main() {
	size := 300

	solverR := propagator.NewSolver(
		propagator.WithSeed(time.Now().UnixMicro()),
		propagator.SelectDomainsAtRandom(),
	)
	solverG := propagator.NewSolver(
		propagator.WithSeed(time.Now().UnixMicro()+42),
		propagator.SelectDomainsAtRandom(),
	)
	solverB := propagator.NewSolver(
		propagator.WithSeed(time.Now().UnixMicro()+84),
		propagator.SelectDomainsAtRandom(),
	)

	pixelsR := SolvePixelMatrix(size, solverR)
	pixelsG := SolvePixelMatrix(size, solverG)
	pixelsB := SolvePixelMatrix(size, solverB)

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			valueR := pixelsR[x][y].GetAssignedValue()
			valueG := pixelsG[x][y].GetAssignedValue()
			valueB := pixelsB[x][y].GetAssignedValue()
			img.Set(x, y, color.RGBA{R: valueR, G: valueG, B: valueB, A: 255})
		}
	}

	outputFile, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	if err = png.Encode(outputFile, img); err != nil {
		panic(err)
	}
}

func SolvePixelMatrix(size int, solver propagator.Solver) [][]*Pixel {
	values := make([]uint8, 256)
	for i := 0; i < 255; i++ {
		values[i] = uint8(i)
	}

	pixels := make([][]*Pixel, size)

	for i := 0; i < size; i++ {
		pixels[i] = make([]*Pixel, size)
	}

	builder := propagator.BuildModel()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			pixel := propagator.NewVariableFromValues(fmt.Sprintf("%d-%d", x, y), values)
			pixels[x][y] = pixel
			builder.AddDomain(pixel)
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if (x + 1) < size {
				builder.AddConstraint(Adjacency{pixels[x][y], pixels[x+1][y]})
			}
			if (y + 1) < size {
				builder.AddConstraint(Adjacency{pixels[x][y], pixels[x][y+1]})
			}
		}
	}

	model := builder.Build()

	if !solver.Solve(model) {
		panic("unable to solve")
	}

	return pixels
}

type Pixel = propagator.Variable[uint8]

type Adjacency struct {
	P1 *Pixel
	P2 *Pixel
}

func (a Adjacency) Scope() []propagator.Domain {
	return propagator.DomainsOf(a.P1, a.P2)
}

func (a Adjacency) Propagate(m *propagator.Mutator) {
	min1 := math.MaxInt
	max1 := math.MinInt
	for _, s := range a.P1.AllowedValues() {
		if int(s) < min1 {
			min1 = int(s)
		}
		if int(s) > max1 {
			max1 = int(s)
		}
	}

	m.Add(a.P2.ExcludeBy(func(i uint8) bool {
		return int(i) > (max1+10) || int(i) < (min1-10)
	}))

	min2 := math.MaxInt
	max2 := math.MinInt
	for _, s := range a.P2.AllowedValues() {
		if int(s) < min2 {
			min2 = int(s)
		}
		if int(s) > max2 {
			max2 = int(s)
		}
	}

	m.Add(a.P1.ExcludeBy(func(i uint8) bool {
		return int(i) > (max2+10) || int(i) < (min2-10)
	}))
}
