package topic

import (
	"strconv"
	"sync"

	"github.com/andydunstall/figg/service/pkg/commitlog"
)

type Topic struct {
	name string
	// Note choosing a slice over a map. This is since a large majority of
	// accesses is from t.Publish iterating though the subscribers, which is
	// much faster to iterate a slice rather than a map. The cost is
	// unsubscribing becomes O(n) though unsubscribes should be rare and
	// expecting the number of subscribers to be relatively smallk
	subscribers []*Subscription
	segment     *commitlog.Segment
	mu          sync.Mutex
}

func NewTopic(name string) *Topic {
	return &Topic{
		name:        name,
		subscribers: []*Subscription{},
		segment:     commitlog.NewSegment(),
		mu:          sync.Mutex{},
	}
}

func (t *Topic) Name() string {
	return t.name
}

// Offset returns the offset of the last message processed.
func (t *Topic) Offset() uint64 {
	// No need to lock as segment is thread safe.
	return t.segment.Size()
}

// GetMessage returns the message with the given offset. If the offset is
// less than the earliest message, will round up to the next message.
func (t *Topic) GetMessage(offset uint64) ([]byte, uint64, bool) {
	// No need to lock as segment is thread safe.
	return t.segment.Lookup(offset)
}

func (t *Topic) Publish(b []byte) {
	// No need to lock as segment is thread safe.
	offset := t.segment.Append(b)
	serial := strconv.FormatUint(offset, 10)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Notify all subscribers to wake up and send the latest message.
	for _, sub := range t.subscribers {
		sub.Notify(t.name, serial, b)
	}
}

func (t *Topic) Subscribe(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = append(t.subscribers, s)
}

func (t *Topic) SubscribeIfLatest(offset uint64, s *Subscription) bool {
	// No need to lock as segment is thread safe.
	if offset != t.segment.Size() {
		return false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = append(t.subscribers, s)
	return true
}

func (t *Topic) Unsubscribe(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subscribers := make([]*Subscription, 0, len(t.subscribers))
	for _, sub := range t.subscribers {
		if s != sub {
			subscribers = append(subscribers, s)
		}
	}
	t.subscribers = subscribers
}
