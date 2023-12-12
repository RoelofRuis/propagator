package propagator

import (
	"fmt"
	"testing"
)

func TestMutation(t *testing.T) {
	indices := []DomainValue[int]{
		{0, 1.0, 1},
		{0, 0.5, 2},
	}

	domain := NewVariable("test", indices)

	mutator := NewMutator()

	mutator.Add(domain.Exclude(0))

	if !domain.IsUnassigned() {
		t.Errorf("expected free domain before mutation")
	}
	if domain.version != 1 {
		fmt.Printf("%v\n", domain.version)
		t.Errorf("expected version to be 1 before mutation")
	}

	mutator.apply()

	if !domain.IsAssigned() {
		t.Errorf("expected fixed domain after mutation")
	}
	if domain.version != 2 {
		t.Errorf("expected version to be 2 after mutation")
	}

	mutator.revertAll()

	if !domain.IsUnassigned() {
		t.Errorf("expected free domain after revert")
	}
	if domain.version != 3 {
		t.Errorf("expected version to be 3 after revert")
	}
}
