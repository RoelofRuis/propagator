package propagator

import (
	"hash/maphash"
	"math/rand"
)

// Solver is responsible for solving a given model.
type Solver struct {
	rnd            *rand.Rand
	domainPicker   domainPicker
	indexPicker    indexPicker
	maxSolutions   int
	solutionsFound int

	queue  *SetQueue[*Domain]
	events *PubSub
}

// SolverEvent is used as key to hook functions to the solver.
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

// NewSolver creates a new solver. It allows for SolverOptions to customize the solver behavior.
func NewSolver(options ...SolverOption) Solver {
	solver := Solver{
		rnd:            rand.New(rand.NewSource(int64(new(maphash.Hash).Sum64()))),
		domainPicker:   &MinRemainingValuesPicker{},
		indexPicker:    &ProbabilisticIndexPicker{},
		solutionsFound: 0,
		maxSolutions:   1,
		events:         NewPubsub(),
		queue:          NewSetQueue[*Domain](), // domain ids
	}
	for _, opt := range options {
		opt(&solver)
	}
	return solver
}

// Solve runs the solving algorithm on the Model and returns whether a solution could be found.
// The model is updated to reflect the found solution.
func (s *Solver) Solve(model Model) bool {
	s.events.Publish(Start)

	s.domainPicker.init(model, s.rnd)
	s.indexPicker.init(model, s.rnd)

	mutations, success := s.propagate(model, model.Domains...)
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

	domain := s.domainPicker.nextDomain(model)

	if domain == nil {
		return false
	}

	selectMutations := newMutator()

	for {
		selectedIndex := s.indexPicker.nextIndex(domain)
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

func (s *Solver) propagate(model Model, domains ...*Domain) (*Mutator, bool) {
	s.events.Publish(PropagateStart)
	for _, domain := range domains {
		s.queue.Enqueue(domain)
	}
	mutator := newMutator()

	for {
		if s.queue.IsEmpty() {
			return mutator, true
		}
		s.events.Publish(PropagateRound)

		selectedDomain := s.queue.Dequeue()
		targetDomains := Set[*Domain]{}

		for _, constraintId := range model.domainConstraints[selectedDomain.id] {
			constraint := model.constraints[constraintId]

			mutator.setActiveConstraintId(constraintId)
			constraint.constraint.Propagate(mutator)

			for _, targetDomainId := range constraint.linkedDomains {
				targetDomains = targetDomains.Insert(model.Domains[targetDomainId])
			}
		}

		versions := make(map[DomainId]int)
		for targetDomain := range targetDomains {
			versions[targetDomain.id] = targetDomain.version()
		}

		mutator.apply()

		for targetDomain := range targetDomains {
			if targetDomain.IsInContradiction() {
				s.queue.Reset()
				return mutator, false
			}

			if targetDomain.version() > versions[targetDomain.id] {
				s.queue.Enqueue(targetDomain)
			}
		}
	}
}
