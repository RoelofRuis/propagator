package propagator

import (
	"math"
	"testing"
)

func TestDomainStates_Empty(t *testing.T) {
	domain := NewDomain("test", []index{})
	if !domain.IsContradiction() {
		t.Fatalf("empty domain should be in contradiction")
	}
	if domain.IsFixed() {
		t.Fatalf("empty domain should not be fixed")
	}
	if domain.IsFree() {
		t.Fatalf("empty domain should not be free")
	}
}

func TestDomainStates_Fixed(t *testing.T) {
	domain := NewDomain("test", []index{newIndex(1.0, 0)})
	if domain.IsContradiction() {
		t.Fatalf("fixed domain should not be in contradiction")
	}
	if !domain.IsFixed() {
		t.Fatalf("fixed domain should be fixed")
	}
	if domain.IsFree() {
		t.Fatalf("fixed domain should not be free")
	}
}

func TestDomainStates_Free(t *testing.T) {
	domain := NewDomain("test", []index{newIndex(1.0, 0), newIndex(1.0, 0)})
	if domain.IsContradiction() {
		t.Fatalf("free domain should not be in contradiction")
	}
	if domain.IsFixed() {
		t.Fatalf("free domain should not be fixed")
	}
	if !domain.IsFree() {
		t.Fatalf("free domain should be free")
	}
}

func TestDomainStates_Contradict(t *testing.T) {
	domain := NewDomain("test", []index{newIndex(0.0, 0), newIndex(0.0, 0)})
	if !domain.IsContradiction() {
		t.Fatalf("contradicted domain should be in contradiction")
	}
	if domain.IsFixed() {
		t.Fatalf("contradicted domain should not be fixed")
	}
	if domain.IsFree() {
		t.Fatalf("contradicted domain should not be free")
	}
}

func TestEntropyAndPriority(t *testing.T) {
	type test struct {
		domain              *Domain
		expectedEntropy     float64
		expectedMinPriority int
	}

	tests := []test{
		{
			NewDomain("test", []index{}),
			math.Inf(-1),
			math.MaxInt,
		},
		{
			NewDomain("test", []index{newIndex(1.0, 0)}),
			0.0,
			0,
		},
		{
			NewDomain("test", []index{newIndex(1.0, 0), newIndex(1.0, 0)}),
			1.0,
			0,
		},
		{
			NewDomain("test", []index{newIndex(1.0, 1), newIndex(1.0, 1)}),
			1.0,
			1,
		},
		{
			NewDomain("test", []index{newIndex(4.0, 0), newIndex(1.0, 0)}),
			0.7219280948,
			0,
		},
		{
			NewDomain("test", []index{newIndex(1.0, 0), newIndex(1.0, 0), newIndex(1.0, 0), newIndex(1.0, 0)}),
			2.0,
			0,
		},
		{
			NewDomain("test", []index{newIndex(1.0, 0), newIndex(1.0, 1)}),
			0.0,
			0,
		},
		{
			NewDomain("test", []index{newIndex(1.0, 0), newIndex(1.0, 0), newIndex(1.0, 1)}),
			1.0,
			0,
		},
	}

	for _, tc := range tests {
		gotEntropy := tc.domain.Entropy()
		if math.Abs(gotEntropy-tc.expectedEntropy) > 1e-10 {
			t.Fatalf("ENTROPY expected %v, got: %v", tc.expectedEntropy, gotEntropy)
		}
		gotPriority := tc.domain.MinPriority()
		if gotPriority != tc.expectedMinPriority {
			t.Fatalf("MIN PRIORITY expected %v, got %v", tc.expectedMinPriority, gotPriority)
		}
	}
}
