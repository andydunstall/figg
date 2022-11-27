package topic

import (
	"strconv"
	"sync"
)

type Topic struct {
	name        string
	subscribers map[*Subscription]interface{}
	messages    map[uint64][]byte
	offset      uint64
	mu          sync.Mutex
}

func NewTopic(name string) *Topic {
	return &Topic{
		name:        name,
		subscribers: map[*Subscription]interface{}{},
		messages:    map[uint64][]byte{},
		offset:      0,
		mu:          sync.Mutex{},
	}
}

func (t *Topic) Name() string {
	return t.name
}

// Offset returns the offset of the last message processed.
func (t *Topic) Offset() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.offset
}

// GetMessage returns the message with the given offset. If the offset is
// less than the earliest message, will round up to the next message.
func (t *Topic) GetMessage(offset uint64) ([]byte, uint64, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

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

	serial := strconv.FormatUint(t.offset, 10)
	// Notify all subscribers to wake up and send the latest message.
	for sub, _ := range t.subscribers {
		sub.Notify(serial, b)
	}
}

func (t *Topic) Subscribe(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[s] = struct{}{}
}

func (t *Topic) SubscribeIfLatest(offset uint64, s *Subscription) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if offset != t.offset {
		return false
	}

	t.subscribers[s] = struct{}{}
	return true
}

func (t *Topic) Unsubscribe(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, s)
}
