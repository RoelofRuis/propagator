package propagator

import (
	"fmt"
	"testing"
)

var table = []struct {
	numInts int
}{
	{numInts: 10},
	{numInts: 100},
	{numInts: 1000},
	{numInts: 10000},
}

func BenchmarkVariable_ExcludeBy(b *testing.B) {
	for _, v := range table {
		b.Run(fmt.Sprintf("domain_size_%d", v.numInts), func(b *testing.B) {
			ints := make([]int, v.numInts)
			for i := 0; i < v.numInts; i++ {
				ints[i] = i
			}

			problem := NewProblem()
			variable := AddVariable(problem, "test", AsDomainValues(ints...))
			problem.Model()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				variable.ExcludeBy(func(i int) bool {
					return i%2 == 0
				})
			}
		})
	}
}
