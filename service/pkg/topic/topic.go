package topic

import (
	"strconv"
	"sync"
)

type Topic struct {
	name string
	// Note choosing a slice over a map. This is since a large majority of
	// accesses is from t.Publish iterating though the subscribers, which is
	// much faster to iterate a slice rather than a map. The cost is
	// unsubscribing becomes O(n) though unsubscribes should be rare and
	// expecting the number of subscribers to be relatively smallk
	subscribers []*Subscription
	messages    map[uint64][]byte
	offset      uint64
	mu          sync.Mutex
}

func NewTopic(name string) *Topic {
	return &Topic{
		name:        name,
		subscribers: []*Subscription{},
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
	for _, sub := range t.subscribers {
		sub.Notify(serial, b)
	}
}

func (t *Topic) Subscribe(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = append(t.subscribers, s)
}

func (t *Topic) SubscribeIfLatest(offset uint64, s *Subscription) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if offset != t.offset {
		return false
	}

	t.subscribers = append(t.subscribers, s)
	return true
}

func (t *Topic) Unsubscribe(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subscribers := make([]*Subscription, len(t.subscribers))
	for _, sub := range t.subscribers {
		if s != sub {
			subscribers = append(subscribers, s)
		}
	}
	t.subscribers = subscribers
}
