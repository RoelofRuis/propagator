package propagator

import (
	"testing"
)

func TestPubSub_Publish(t *testing.T) {
	ps := NewPubsub()

	// Test publishing to a key with subscribed callbacks
	called := false
	callback := func() { called = true }
	ps.Subscribe("key1", callback)

	ps.Publish("key1")

	if !called {
		t.Error("Expected callback to be called")
	}
}

func TestPubSub_Publish_NoCallbacks(t *testing.T) {
	ps := NewPubsub()

	// Test publishing to a key with no subscribed callbacks
	ps.Publish("key1")

	// No callbacks are expected, so no assertions are necessary
}
