package propagator

import (
	"testing"
)

func TestCreate(t *testing.T) {
	i1 := indexFactorySingleton.create(1.0, 0)
	i2 := indexFactorySingleton.create(1.0, 0)
	i3 := indexFactorySingleton.create(2.0, 0)

	if i1 != i2 {
		t.Fatalf("similar indices should have same reference")
	}

	if i1 == i3 {
		t.Fatalf("non-similar indices should have different reference")
	}
}

func TestNewIndex(t *testing.T) {
	i := indexFactorySingleton.create(1.0, 0)

	probability, priority := unpackPriorityProbability(i.probAndPrio)

	if probability != 1.0 {
		t.Fatalf("probability should be 1.0")
	}

	if priority != 0 {
		t.Fatalf("priority should be 0")
	}
}

func TestAdjustProbability(t *testing.T) {
	i := indexFactorySingleton.create(1.0, 0)

	i2, success := i.adjust(-1, 0.5, 0)
	if !success {
		t.Fatalf("index probability should be decremented")
	}

	_, success = i2.adjust(-1, 1.0, 0)
	if success {
		t.Fatalf("index probability should not be incremented")
	}
}

func TestAdjustPriority(t *testing.T) {
	i := indexFactorySingleton.create(1.0, 0)

	i2, success := i.adjust(-1, 1.0, 1)
	if !success {
		t.Fatalf("index priority should be incremented")
	}

	_, success = i2.adjust(-1, 1.0, 0)
	if success {
		t.Fatalf("index priority should not be decremented")
	}
}
