package propagator

import (
	"testing"
)

func TestSetIntersect(t *testing.T) {
	// Create sets for testing
	set1 := NewSet(1, 2, 3, 4)
	set2 := NewSet(3, 4, 5, 6)

	// Calculate the intersection
	intersection := set1.Intersect(set2)

	// Check the expected result
	expected := NewSet(3, 4)
	if !setsAreEqual(intersection, expected) {
		t.Errorf("Intersection is incorrect. Expected: %v, Got: %v", expected, intersection)
	}
}

func TestSetInsert(t *testing.T) {
	// Create an empty set
	set := Set[int]{} // Update with the appropriate type

	// Insert a value into the set
	set = set.Insert(42)

	// Check if the set contains the inserted value
	if !set.Contains(42) {
		t.Errorf("Value not inserted into the set: %v", set)
	}
}

func TestSetContains(t *testing.T) {
	// Create a set for testing
	set := NewSet("apple", "banana", "orange")

	// Check if the set contains a specific value
	if !set.Contains("banana") {
		t.Errorf("Set does not contain expected value")
	}
}

func TestSetContainsOneOf(t *testing.T) {
	// Create a set for testing
	set := NewSet("apple", "banana", "orange")

	// Check if the set contains any of the provided values
	values := []string{"banana", "grape", "watermelon"}
	if !set.ContainsOneOf(values) {
		t.Errorf("Set does not contain any of the expected values")
	}
}

func TestSetSize(t *testing.T) {
	// Create a set for testing
	set := NewSet(1, 2, 3, 4, 5)

	// Check the size of the set
	expectedSize := 5
	if setSize := set.Size(); setSize != expectedSize {
		t.Errorf("Set size is incorrect. Expected: %d, Got: %d", expectedSize, setSize)
	}
}

func TestSetEmpty(t *testing.T) {
	// Create an empty set
	emptySet := Set[int]{} // Update with the appropriate type

	// Check if the set is empty
	if !emptySet.IsEmpty() {
		t.Errorf("IsEmpty set is not recognized as empty")
	}

	// Create a non-empty set
	nonEmptySet := NewSet("apple", "banana")

	// Check if the set is empty
	if nonEmptySet.IsEmpty() {
		t.Errorf("Non-empty set is recognized as empty")
	}
}

// Helper function to check equality of sets
func setsAreEqual(set1, set2 Set[int]) bool {
	if set1.Size() != set2.Size() {
		return false
	}

	for elem := range set1 {
		if !set2.Contains(elem) {
			return false
		}
	}

	return true
}
