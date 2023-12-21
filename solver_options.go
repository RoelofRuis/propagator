package propagator

import (
	"log"
	"math/rand"
	"reflect"
	"strings"
)

// SolverOption functional option for the Solver.
type SolverOption func(solver *Solver)

// WithSeed explicitly sets the random seed to allow reproducible randomness.
func WithSeed(int int64) SolverOption {
	return func(s *Solver) {
		s.rnd = rand.New(rand.NewSource(int))
	}
}

// On hooks a function to the solver when the given SolverEvent fires.
func On(event SolverEvent, f func()) SolverOption {
	return func(s *Solver) {
		s.events.Subscribe(event, f)
	}
}

// LogInfo logs solver info during running.
func LogInfo() SolverOption {
	return func(s *Solver) {
		round := 0
		s.events.Subscribe(Start, func() { log.Printf("[SOLVER] Starting\n") })
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

// LogConstraints logs the model constraints when solving is started.
func LogConstraints(model Model) SolverOption {
	return func(s *Solver) {
		s.events.Subscribe(Start, func() {
			log.Printf("CONSTRAINTS:\n")
			for i, boundConstraint := range model.constraints {
				constraintName := reflect.TypeOf(boundConstraint.constraint)
				var links []string
				for _, domain := range boundConstraint.linkedDomains {
					links = append(links, model.Domains[domain].Name())
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

// SelectIndicesAtRandom selects next indices at random from the available indices.
// This picker does not take into account any probability and priority values.
func SelectIndicesAtRandom() SolverOption {
	return SelectIndicesBy(&RandomIndexPicker{})
}

// SelectIndicesProbabilistically selects next indices based on chance using the index probabilities. It also takes into
// account the priority values, selecting only from the group indices with the lowest priority value.
func SelectIndicesProbabilistically() SolverOption {
	return SelectIndicesBy(&ProbabilisticIndexPicker{})
}

// SelectIndicesBy sets the index picker.
func SelectIndicesBy(picker indexPicker) SolverOption {
	return func(s *Solver) {
		s.indexPicker = picker
	}
}

// SelectDomainsByIndex select next Domains in order by index.
func SelectDomainsByIndex() SolverOption {
	return SelectDomainsBy(&IndexDomainPicker{})
}

// SelectDomainsAtRandom selects next Domains at random.
func SelectDomainsAtRandom() SolverOption {
	return SelectDomainsBy(&RandomDomainPicker{})
}

// SelectDomainsByMinRemainingValues selects the next Domain with the minimum number of remaining free values.
func SelectDomainsByMinRemainingValues() SolverOption {
	return SelectDomainsBy(&MinRemainingValuesPicker{})
}

// SelectDomainsByMinEntropy selects next Domain by minimal Shannon entropy.
func SelectDomainsByMinEntropy() SolverOption {
	return SelectDomainsBy(&MinEntropyDomainPicker{})
}

// SelectDomainsBy sets the domain picker.
func SelectDomainsBy(picker domainPicker) SolverOption {
	return func(s *Solver) {
		s.domainPicker = picker
	}
}
