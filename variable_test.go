package propagator

import (
	"math"
	"testing"
)

func TestDomainGetName(t *testing.T) {
	domain := NewVariable[int]("test", nil)
	if domain.GetName() != "test" {
		t.Fatalf("invalid name returned")
	}
}

func TestDomainEmpty(t *testing.T) {
	domain := NewVariable[int]("test", nil)
	if !domain.IsInContradiction() {
		t.Fatalf("empty domain should be in contradiction")
	}
	if domain.IsAssigned() {
		t.Fatalf("empty domain should not be fixed")
	}
	if domain.IsUnassigned() {
		t.Fatalf("empty domain should not be free")
	}
}

func TestDomainAssigned(t *testing.T) {
	domain := NewVariable[int]("test", []DomainValue[int]{{0, 1, 1}})
	if domain.IsInContradiction() {
		t.Fatalf("fixed domain should not be in contradiction")
	}
	if !domain.IsAssigned() {
		t.Fatalf("fixed domain should be fixed")
	}
	if domain.IsUnassigned() {
		t.Fatalf("fixed domain should not be free")
	}
}

func TestDomainUnassigned(t *testing.T) {
	domain := NewVariable("test", []DomainValue[int]{{0, 1, 1}, {0, 1, 2}})
	if domain.IsInContradiction() {
		t.Fatalf("free domain should not be in contradiction")
	}
	if domain.IsAssigned() {
		t.Fatalf("free domain should not be fixed")
	}
	if !domain.IsUnassigned() {
		t.Fatalf("free domain should be free")
	}
}

func TestDomainContradicts(t *testing.T) {
	domain := NewVariable("test", []DomainValue[int]{{0, 0, 1}})
	if !domain.IsInContradiction() {
		t.Fatalf("contradicted domain should be in contradiction")
	}
	if domain.IsAssigned() {
		t.Fatalf("contradicted domain should not be fixed")
	}
	if domain.IsUnassigned() {
		t.Fatalf("contradicted domain should not be free")
	}
}

func TestEntropyAndPriority(t *testing.T) {
	type test struct {
		variable            *Variable[int]
		expectedEntropy     float64
		expectedMinPriority int
	}

	tests := []test{
		{
			NewVariable("test", []DomainValue[int]{}),
			math.Inf(-1),
			math.MaxInt,
		},
		{
			NewVariable("test", []DomainValue[int]{{0, 1.0, 1}}),
			0.0,
			0,
		},
		{
			NewVariable("test", []DomainValue[int]{{0, 1.0, 1}, {0, 1.0, 2}}),
			1.0,
			0,
		},
		{
			NewVariable("test", []DomainValue[int]{{1, 1.0, 1}, {1, 1.0, 2}}),
			1.0,
			1,
		},
		{
			NewVariable("test", []DomainValue[int]{{0, 4.0, 1}, {0, 1.0, 2}}),
			0.7219280948,
			0,
		},
		{
			NewVariable("test", []DomainValue[int]{{0, 1.0, 1}, {0, 1.0, 2}, {0, 1.0, 3}, {0, 1.0, 4}}),
			2.0,
			0,
		},
		{
			NewVariable("test", []DomainValue[int]{{0, 1.0, 1}, {1, 1.0, 2}}),
			0.0,
			0,
		},
		{
			NewVariable("test", []DomainValue[int]{{0, 1.0, 1}, {0, 1.0, 2}, {1, 1.0, 3}}),
			1.0,
			0,
		},
	}

	for _, tc := range tests {
		gotEntropy := tc.variable.Entropy()
		if math.Abs(gotEntropy-tc.expectedEntropy) > 1e-10 {
			t.Fatalf("ENTROPY expected %v, got: %v", tc.expectedEntropy, gotEntropy)
		}
		gotPriority := tc.variable.minPriority
		if gotPriority != tc.expectedMinPriority {
			t.Fatalf("MIN PRIORITY expected %v, got %v", tc.expectedMinPriority, gotPriority)
		}
	}
}
