package main

import (
	"github.com/RoelofRuis/propagator"
	"testing"
)

func BenchmarkSolve(b *testing.B) {
	solver := propagator.NewSolver(
		propagator.WithSeed(0),
	)

	for i := 0; i < b.N; i++ {
		Solve(solver)
	}
}
