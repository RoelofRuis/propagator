package propagator

// PubSub is a very simple publish/subscribe wrapper where events are just registered callbacks.
type PubSub struct {
	subscriptions map[string][]func()
}

func NewPubsub() *PubSub {
	return &PubSub{
		subscriptions: make(map[string][]func()),
	}
}

func (e *PubSub) Subscribe(key string, callback func()) {
	e.subscriptions[key] = append(e.subscriptions[key], callback)
}

func (e *PubSub) Publish(key string) {
	for _, callback := range e.subscriptions[key] {
		callback()
	}
}
