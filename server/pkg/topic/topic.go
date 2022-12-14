package topic

import (
	"sync"

	"github.com/andydunstall/figg/server/pkg/commitlog"
)

type Message struct {
	Topic   string
	Message []byte
	Offset  uint64
}

type Topic struct {
	name string
	log  *commitlog.CommitLog

	// Mutex protecting the below fields.
	mu sync.Mutex

	// Note choosing a slice over a map. This is since a large majority of
	// accesses is from t.Publish iterating though the subscribers, which is
	// much faster to iterate a slice rather than a map. The cost is
	// unsubscribing becomes O(n) though unsubscribes should be rare and
	// expecting the number of subscribers to be relatively smallk
	subscribers []*Subscription
	offset      uint64
}

func NewTopic(name string, options Options) *Topic {
	log := commitlog.NewCommitLog(
		options.Persisted,
		options.SegmentSize,
		options.Dir+"/"+name,
	)
	return &Topic{
		name:        name,
		log:         log,
		mu:          sync.Mutex{},
		subscribers: []*Subscription{},
		offset:      0,
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
func (t *Topic) GetMessage(offset uint64) ([]byte, error) {
	b, err := t.log.Lookup(offset)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (t *Topic) Publish(b []byte) {
	// Add to the commit log before sending to subscribers.
	t.log.Append(b)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.offset += uint64(len(b) + 4)

	// Notify all subscribers to wake up and send the latest message.
	m := Message{
		Topic:   t.name,
		Message: b,
		Offset:  t.offset,
	}
	for _, sub := range t.subscribers {
		sub.Notify(m)
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

	subscribers := make([]*Subscription, 0, len(t.subscribers))
	for _, sub := range t.subscribers {
		if s != sub {
			subscribers = append(subscribers, s)
		}
	}
	t.subscribers = subscribers
}
