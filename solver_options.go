package propagator

import (
	"log"
	"reflect"
	"strings"
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

func LogConstraints(model Model) SolverOption {
	return func(s *Solver) {
		s.events.Subscribe(Start, func() {
			log.Printf("CONSTRAINTS:\n")
			for i, boundConstraint := range model.constraints {
				constraintName := reflect.TypeOf(boundConstraint.constraint)
				var links []string
				for _, domain := range boundConstraint.linkedDomains {
					links = append(links, model.domains[domain].name())
				}
				log.Printf("%-4d %s\n     %s\n", i, constraintName, strings.Join(links, " "))
			}
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
	return SelectDomainsBy(nextDomainByIndex)
}

// SelectDomainsAtRandom selects next domains at random.
func SelectDomainsAtRandom() SolverOption {
	return SelectDomainsBy(nextDomainAtRandom)
}

// SelectDomainsBy sets the domain picker function.
func SelectDomainsBy(picker domainPicker) SolverOption {
	return func(s *Solver) {
		s.domainPicker = picker
	}
}
