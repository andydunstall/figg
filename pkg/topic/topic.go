package topic

import (
	"sync"
)

type Subscription interface {
	Signal()
}

type Topic struct {
	subscribers map[Subscription]interface{}
	messages    map[uint64][]byte
	offset      uint64
	mu          sync.RWMutex
}

func NewTopic() *Topic {
	return &Topic{
		subscribers: map[Subscription]interface{}{},
		messages:    map[uint64][]byte{},
		offset:      0,
		mu:          sync.RWMutex{},
	}
}

// Offset returns the offset of the last message processed.
func (t *Topic) Offset() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.offset
}

// GetMessage returns the message with the given offset. If the offset is
// less than the earliest message, will round up to the next message.
func (t *Topic) GetMessage(offset uint64) ([]byte, uint64, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if offset > t.offset {
		return nil, 0, false
	}
	m, ok := t.messages[offset]
	if !ok {
		return nil, 0, false
	}
	return m, offset, true
}

func (t *Topic) Publish(b []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.offset += 1
	t.messages[t.offset] = b

	// Notify all subscribers to wake up and send the latest message.
	for sub, _ := range t.subscribers {
		sub.Signal()
	}
}

func (t *Topic) Subscribe(s Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[s] = struct{}{}
}

func (t *Topic) Unsubscribe(s Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, s)
}
