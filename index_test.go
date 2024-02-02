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
	tests := []struct {
		description     string
		index           *index
		adjustProb      Probability
		adjustPrio      Priority
		expectedSuccess bool
		expectedProb    Probability
		expectedPrio    Priority
	}{
		{
			"update nothing",
			indexFactorySingleton.create(1.0, 0),
			1.0,
			0,
			false,
			1.0,
			0,
		},
		{
			"reduce probability",
			indexFactorySingleton.create(1.0, 0),
			0.5,
			0,
			true,
			0.5,
			0,
		},
		{
			"reduce probability with adjusted prio",
			indexFactorySingleton.create(1.0, 1),
			0.5,
			0,
			true,
			0.5,
			1,
		},
		{
			"increase probability fails",
			indexFactorySingleton.create(0.5, 0),
			1.0,
			0,
			false,
			1.0,
			0,
		},
		{
			"increase priority",
			indexFactorySingleton.create(1.0, 0),
			1.0,
			1,
			true,
			1.0,
			1,
		},
		{
			"increase priority with adjusted prob",
			indexFactorySingleton.create(0.5, 0),
			1.0,
			1,
			true,
			0.5,
			1,
		},
		{
			"decrease priority fails",
			indexFactorySingleton.create(1.0, 1),
			1.0,
			0,
			false,
			1.0,
			1,
		},
		{
			"adjust to equal priority fails",
			indexFactorySingleton.create(1.0, 1),
			1.0,
			1,
			false,
			1.0,
			1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			adjusted, success := tc.index.adjust(-1, tc.adjustProb, tc.adjustPrio)
			if success != tc.expectedSuccess {
				t.Fatalf("expected adjust to have success is %v but got %v", tc.expectedSuccess, success)
			}
			if tc.expectedSuccess == false {
				return
			}

			probability, priority := unpackPriorityProbability(adjusted.probAndPrio)
			if probability != tc.expectedProb {
				t.Fatalf("expected probability '%f' but got '%f'", tc.expectedProb, probability)
			}

			if priority != tc.expectedPrio {
				t.Fatalf("expected priority '%d' but got '%d'", tc.expectedPrio, priority)
			}
		})
	}
}
