package main

import (
	"fmt"
	"github.com/RoelofRuis/propagator"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
)

func main() {
	size := 100

	pixels := SolvePixelMatrix(size)

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			value := pixels[x][y].GetFixedValue()
			img.Set(x, y, color.RGBA{R: value, G: value, B: value, A: 255})
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

func SolvePixelMatrix(size int) [][]*Pixel {
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
			builder.AddDomain(pixel.Domain)
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

	solver := propagator.NewSolver(model, propagator.WithSeed(time.Now().UnixMicro()))

	if !solver.Solve() {
		panic("unable to solve")
	}

	return pixels
}

type Pixel = propagator.Variable[uint8]

type Adjacency struct {
	P1 *Pixel
	P2 *Pixel
}

func (a Adjacency) Scope() []*propagator.Domain {
	return []*propagator.Domain{a.P1.Domain, a.P2.Domain}
}

func (a Adjacency) Propagate(m *propagator.Mutator) {
	if a.P1.IsFixed() && a.P2.IsFree() {
		fv := a.P1.GetFixedValue()
		m.Add(a.P2.BanBy(func(i uint8) bool {
			return i > (fv+10) || i < (fv-10)
		}))
	}
	if a.P2.IsFixed() && a.P1.IsFree() {
		fv := a.P2.GetFixedValue()
		m.Add(a.P1.BanBy(func(i uint8) bool {
			return i > (fv+10) || i < (fv-10)
		}))
	}
}
