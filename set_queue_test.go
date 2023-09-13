package propagator

import (
	"testing"
)

func TestSetQueueEnqueue(t *testing.T) {
	queue := NewSetQueue[int]()
	queue.Enqueue(1, 2, 3)

	// Enqueue a new element
	queue.Enqueue(4)

	// Check if the element is added to the queue
	if !queueContains(queue, 4) {
		t.Errorf("Enqueued element not found in the queue: %v", queue)
	}

	// Enqueue an existing element
	queue.Enqueue(3)

	// Check if the existing element is not added to the queue again
	if countOccurrences(queue.Elements, 3) > 1 {
		t.Errorf("Existing element added multiple times to the queue: %v", queue)
	}
}

func TestSetQueueIsEmpty(t *testing.T) {
	emptyQueue := SetQueue[int]{} // Update with the appropriate type

	// Check if the empty queue is recognized as empty
	if !emptyQueue.IsEmpty() {
		t.Errorf("IsEmpty queue is not recognized as empty")
	}

	nonEmptyQueue := NewSetQueue[int]()
	nonEmptyQueue.Enqueue(1, 2, 3)

	// Check if the non-empty queue is recognized as non-empty
	if nonEmptyQueue.IsEmpty() {
		t.Errorf("Non-empty queue is recognized as empty")
	}
}

func TestSetQueueDequeue(t *testing.T) {
	queue := NewSetQueue[int]()
	queue.Enqueue(1, 2, 3)

	// Dequeue an element
	dequeued := queue.Dequeue()

	// Check if the correct element is dequeued
	expected := 1
	if dequeued != expected {
		t.Errorf("Dequeued element is incorrect. Expected: %v, Got: %v", expected, dequeued)
	}

	// Check if the dequeued element is removed from the set
	if queue.set.Contains(dequeued) {
		t.Errorf("Dequeued element still exists in the set: %v", queue)
	}

	// Check if the dequeued element is removed from the queue
	if queueContains(queue, dequeued) {
		t.Errorf("Dequeued element still exists in the queue: %v", queue)
	}
}

// Helper function to check if the queue contains an element
func queueContains(queue *SetQueue[int], elem int) bool {
	for _, e := range queue.Elements {
		if e == elem {
			return true
		}
	}
	return false
}

// Helper function to count occurrences of an element in a slice
func countOccurrences(slice []int, elem int) int {
	count := 0
	for _, e := range slice {
		if e == elem {
			count++
		}
	}
	return count
}
