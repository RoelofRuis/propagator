package main

import "testing"

func BenchmarkSolvePixelMatrix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SolvePixelMatrix(32, 0)
	}
}
