package main

import (
	"github.com/RoelofRuis/propagator"
	"testing"
)

func BenchmarkSolvePixelMatrix(b *testing.B) {
	solver := propagator.NewSolver(
		propagator.WithSeed(0),
	)

	for i := 0; i < b.N; i++ {
		SolvePixelMatrix(32, solver)
	}
}
