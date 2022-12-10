package figg

import (
	"sync"
)

type AttachQueue struct {
	pending map[string][]func()
	mu      sync.Mutex
}

func NewAttachQueue() *AttachQueue {
	return &AttachQueue{
		pending: make(map[string][]func()),
		mu:      sync.Mutex{},
	}
}

func (q *AttachQueue) Push(topic string, cb func()) {
	q.mu.Lock()
	defer q.mu.Unlock()

	pending, ok := q.pending[topic]
	if !ok {
		pending = []func(){}
	}

	pending = append(pending, cb)
	q.pending[topic] = pending
}

func (q *AttachQueue) Attached(topic string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	pending, ok := q.pending[topic]
	if !ok {
		return
	}

	for _, cb := range pending {
		cb()
	}
	delete(q.pending, topic)
}
