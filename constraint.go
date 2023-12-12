package propagator

// Constraint describes the way domains depend on each other and allows for propagating updated values.
type Constraint interface {
	// Scope returns all domains that are influenced by this constraint.
	Scope() []DomainId
	// Propagate is called every time a domain in the constraint scope is updated and allows for further updates to be passed.
	Propagate(m *Mutator)
}

type constraintId = int
