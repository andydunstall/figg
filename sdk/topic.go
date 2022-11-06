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

func (t *Topic) Subscribe(s MessageSubscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[s] = struct{}{}
}

func (t *Topic) Unsubscribe(s MessageSubscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, s)
}
