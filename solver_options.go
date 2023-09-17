package propagator

import (
	"log"
)

type SolverOption func(solver *Solver)

func WithSeed(int int64) SolverOption {
	return func(s *Solver) {
		s.rndSeed = int
	}
}

func On(event SolverEvent, f func()) SolverOption {
	return func(s *Solver) {
		s.events.Subscribe(event, f)
	}
}

func LogInfo() SolverOption {
	return func(s *Solver) {
		round := 0
		s.events.Subscribe(Start, func() { log.Printf("[SOLVER] Starting\nRND Seed [%d]\n", s.rndSeed) })
		s.events.Subscribe(Failure, func() { log.Print("[SOLVER] Failed finding solution\n") })
		s.events.Subscribe(SolutionFound, func() { log.Print("[SOLVER] Solution found\n") })
		s.events.Subscribe(Select, func() {
			log.Printf("[SOLVER] Next round (%d)\n", round)
			round++
		})
		s.events.Subscribe(PropagateStart, func() {
			log.Printf("[SOLVER] Start propagating constraints\n")
		})
		s.events.Subscribe(PropagateRound, func() {
			log.Printf("[SOLVER] Propagate round [Queue size %d]\n", s.queue.Length())
		})
	}
}

// FindNSolutions stops the solver after finding a maximum of n solutions.
func FindNSolutions(n int) SolverOption {
	return func(s *Solver) {
		s.maxSolutions = n
	}
}

// FindAllSolutions searches for all solutions exhaustively.
func FindAllSolutions() SolverOption {
	return func(s *Solver) {
		s.maxSolutions = -1
	}
}

// SelectDomainsByIndex select next domains in order by index.
func SelectDomainsByIndex() SolverOption {
	return func(s *Solver) {
		s.nextDomain = nextDomainByIndex
	}
}
