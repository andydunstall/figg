package figg

import (
	"sync"
)

// MessageHandler is a callback function that processes messages delivered to
// subscribers.
type MessageHandler func(topic string, m []byte)

type MessageSubscriber struct {
	CB MessageHandler
}

type Topic struct {
	name        string
	subscribers map[*MessageSubscriber]interface{}
	offset      string
	mu          sync.Mutex
}

func NewTopic(name string) *Topic {
	return &Topic{
		name:        name,
		subscribers: make(map[*MessageSubscriber]interface{}),
		offset:      "",
		mu:          sync.Mutex{},
	}
}

func (t *Topic) OnMessage(m []byte, offset string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.offset = offset

	for sub, _ := range t.subscribers {
		sub.CB(t.name, m)
	}
}

func (t *Topic) Offset() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.offset
}

// Subscribes to the topic. Returns true if this is the first subscriber, false
// otherwise.
func (t *Topic) Subscribe(cb MessageHandler) (*MessageSubscriber, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sub := &MessageSubscriber{
		CB: cb,
	}

	activated := len(t.subscribers) == 0
	t.subscribers[sub] = struct{}{}
	return sub, activated
}

// Unsubscribes from the topic. Returns true if the topic now has no subscribers
// (so is inactive), false otherwise.
func (t *Topic) Unsubscribe(sub *MessageSubscriber) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, sub)

	return len(t.subscribers) == 0
}
