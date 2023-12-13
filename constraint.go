package propagator

// Constraint describes the way Domains depend on each other and allows for propagating updated values.
type Constraint interface {
	// Scope returns all Domains that are influenced by this constraint.
	Scope() []DomainId
	// Propagate is called every time a domain in the constraint scope is updated and allows for further updates to be passed.
	Propagate(m *Mutator)
}

type constraintId = int

// IdsOf extracts the DomainId from a list of variables.
func IdsOf[T comparable](vars ...*Variable[T]) []DomainId {
	domainIds := make([]DomainId, 0, len(vars))
	for _, v := range vars {
		domainIds = append(domainIds, v.Domain.id)
	}
	return domainIds
}
