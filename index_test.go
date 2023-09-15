package propagator

import (
	"testing"
)

func TestNewIndex(t *testing.T) {
	i := newIndex(1.0, 0)
	if i.isBanned {
		t.Fatalf("index should not be banned")
	}
	if i.probability != 1.0 {
		t.Fatalf("index should be 1.0")
	}
}

func TestAdjustProbability(t *testing.T) {
	i := newIndex(1.0, 0)

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
	i := newIndex(1.0, 0)

	i2, success := i.adjust(-1, 1.0, 1)
	if !success {
		t.Fatalf("index priority should be incremented")
	}

	_, success = i2.adjust(-1, 1.0, 0)
	if success {
		t.Fatalf("index priority should not be decremented")
	}
}
