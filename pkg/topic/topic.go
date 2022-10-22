package topic

import (
	"sync"
)

type Topic struct {
	subscribers map[Subscriber]interface{}
	mu          sync.Mutex
}

func NewTopic() *Topic {
	return &Topic{
		subscribers: map[Subscriber]interface{}{},
		mu:          sync.Mutex{},
	}
}

func (t *Topic) Publish(b []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for sub, _ := range t.subscribers {
		sub.Notify(b)
	}
}

func (t *Topic) Subscribe(sub Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[sub] = struct{}{}
}

func (t *Topic) Unsubscribe(sub Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, sub)
}
