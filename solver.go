package propagator

import (
	"math/rand"
)

// Solver is responsible for solving a given model.
type Solver struct {
	model          Model
	rndSeed        int64
	nextDomain     domainPicker
	nextIndex      indexPicker
	maxSolutions   int
	solutionsFound int

	queue  *SetQueue[*Domain]
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

func NewSolver(m Model, options ...SolverOption) Solver {
	solver := Solver{
		model:          m,
		rndSeed:        0,
		nextDomain:     nextDomainByMinEntropy,
		nextIndex:      nextIndexByProbability,
		solutionsFound: 0,
		maxSolutions:   1,
		events:         NewPubsub(),
		queue:          NewSetQueue[*Domain](),
	}
	for _, opt := range options {
		opt(&solver)
	}
	return solver
}

func (s *Solver) Solve() bool {
	rand.Seed(s.rndSeed)

	s.events.Publish(Start)

	mutations, success := s.propagate(s.model.domains...)
	if success {
		s.events.Publish(SearchStart)
		s.selectNext(0)
	}

	hasSolutions := s.solutionsFound > 0
	if !hasSolutions {
		s.events.Publish(Failure)
		mutations.revertAll()
	}

	s.events.Publish(Finished)
	return hasSolutions
}

func (s *Solver) selectNext(level int) bool {
	s.events.Publish(Select)

	if s.isSolved() {
		s.solutionsFound++
		s.events.Publish(SolutionFound)
		if s.maxSolutions > 0 && (s.maxSolutions == s.solutionsFound) {
			return true
		}
	}

	domain := s.nextDomain(s.model)

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

		selectMutations.Add(domain.Fix(selectedIndex))
		selectMutations.apply()

		propagateMutations, success := s.propagate(domain)

		if success && s.selectNext(level+1) {
			return true
		}

		propagateMutations.revertAll()
		selectMutations.revertPrevious()
		selectMutations.Add(domain.Ban(selectedIndex))
		selectMutations.apply()
	}
}

func (s *Solver) propagate(domains ...*Domain) (*Mutator, bool) {
	s.events.Publish(PropagateStart)
	s.queue.Enqueue(domains...)
	mutator := NewMutator()

	for {
		if s.queue.IsEmpty() {
			return mutator, true
		}
		s.events.Publish(PropagateRound)

		selectedDomain := s.queue.Dequeue()
		targetDomains := Set[*Domain]{}

		for _, constraintId := range s.model.domainConstraints[selectedDomain] {
			constraint := s.model.constraints[constraintId]

			mutator.setActiveConstraintId(constraintId)
			constraint.constraint.Propagate(mutator)

			for _, targetDomain := range constraint.linkedDomains {
				targetDomains = targetDomains.Insert(targetDomain)
			}
		}

		versions := make(map[*Domain]int)
		for domain := range targetDomains {
			versions[domain] = domain.version
		}

		mutator.apply()

		for domain := range targetDomains {
			if domain.IsContradiction() {
				s.queue.Reset()
				return mutator, false
			}

			if domain.WasUpdatedSince(versions[domain]) {
				s.queue.Enqueue(domain)
			}
		}
	}
}

func (s *Solver) isSolved() bool {
	for _, domain := range s.model.domains {
		if !domain.IsFixed() {
			return false
		}
	}
	return true
}
