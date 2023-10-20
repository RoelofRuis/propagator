package propagator

import (
	"math/rand"
)

// Solver is responsible for solving a given model.
type Solver struct {
	rndSeed        int64
	nextDomain     domainPicker
	nextIndex      indexPicker
	maxSolutions   int
	solutionsFound int

	queue  *SetQueue[Domain]
	events *PubSub
}

type SolverEvent = string

const (
	Start          SolverEvent = "Start"
	Finished       SolverEvent = "Finished"
	SolutionFound  SolverEvent = "SolutionFound"
	Failure        SolverEvent = "Failure"
	SearchStart    SolverEvent = "SearchStart"
	PropagateStart SolverEvent = "PropagateStart"
	PropagateRound SolverEvent = "PropagateRound"
	Select         SolverEvent = "Select"
)

func NewSolver(options ...SolverOption) Solver {
	solver := Solver{
		rndSeed:        0,
		nextDomain:     nextDomainByMinEntropy,
		nextIndex:      nextIndexByProbability,
		solutionsFound: 0,
		maxSolutions:   1,
		events:         NewPubsub(),
		queue:          NewSetQueue[Domain](), // domain ids
	}
	for _, opt := range options {
		opt(&solver)
	}
	return solver
}

func (s *Solver) Solve(model Model) bool {
	rand.Seed(s.rndSeed)

	s.events.Publish(Start)

	mutations, success := s.propagate(model, model.domains...)
	if success {
		s.events.Publish(SearchStart)
		s.selectNext(0, model)
	}

	hasSolutions := s.solutionsFound > 0
	if !hasSolutions {
		s.events.Publish(Failure)
		mutations.revertAll()
	}

	s.events.Publish(Finished)
	return hasSolutions
}

func (s *Solver) selectNext(level int, model Model) bool {
	s.events.Publish(Select)

	if model.IsSolved() {
		s.solutionsFound++
		s.events.Publish(SolutionFound)
		if s.maxSolutions > 0 && (s.maxSolutions == s.solutionsFound) {
			return true
		}
	}

	domain := s.nextDomain(model)

	if domain == nil {
		return false
	}

	selectMutations := NewMutator()

	for {
		selectedIndex := s.nextIndex(domain)
		if selectedIndex == -1 {
			selectMutations.revertAll()
			return false
		}

		selectMutations.Add(domain.Assign(selectedIndex))
		selectMutations.apply()

		propagateMutations, success := s.propagate(model, domain)

		if success && s.selectNext(level+1, model) {
			return true
		}

		propagateMutations.revertAll()
		selectMutations.revertPrevious()
		selectMutations.Add(domain.Exclude(selectedIndex))
		selectMutations.apply()
	}
}

func (s *Solver) propagate(model Model, domains ...Domain) (*Mutator, bool) {
	s.events.Publish(PropagateStart)
	for _, domain := range domains {
		s.queue.Enqueue(domain)
	}
	mutator := NewMutator()

	for {
		if s.queue.IsEmpty() {
			return mutator, true
		}
		s.events.Publish(PropagateRound)

		selectedDomain := s.queue.Dequeue()
		targetDomains := Set[Domain]{}

		for _, constraintId := range model.domainConstraints[selectedDomain] {
			constraint := model.constraints[constraintId]

			mutator.setActiveConstraintId(constraintId)
			constraint.constraint.Propagate(mutator)

			for _, targetDomain := range constraint.linkedDomains {
				targetDomains = targetDomains.Insert(targetDomain)
			}
		}

		versions := make(map[Domain]int)
		for targetDomain := range targetDomains {
			versions[targetDomain] = targetDomain.getVersion()
		}

		mutator.apply()

		for targetDomain := range targetDomains {
			if targetDomain.IsInContradiction() {
				s.queue.Reset()
				return mutator, false
			}

			if targetDomain.getVersion() > versions[targetDomain] {
				s.queue.Enqueue(targetDomain)
			}
		}
	}
}
