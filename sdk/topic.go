package figg

import (
	"sync"
)

type Topic struct {
	subscribers map[MessageSubscriber]interface{}
	mu          sync.Mutex
}

func NewTopic() *Topic {
	return &Topic{
		subscribers: make(map[MessageSubscriber]interface{}),
		mu:          sync.Mutex{},
	}
}

func (t *Topic) OnMessage(m []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for sub, _ := range t.subscribers {
		sub.NotifyMessage(m)
	}
}

// Subscribes to the topic. Returns true if this is the first subscriber, false
// otherwise.
func (t *Topic) Subscribe(s MessageSubscriber) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	activated := len(t.subscribers) == 0
	t.subscribers[s] = struct{}{}
	return activated
}

func (t *Topic) Unsubscribe(s MessageSubscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, s)
}
