package figg

import (
	"sync"
)

type PendingMessage struct {
	Message *ProtocolMessage
	SeqNum  uint64
}

// MessageQueue contains messages that must be acknowledged but have not been.
type MessageQueue struct {
	pending []PendingMessage
	mu      sync.Mutex
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		pending: []PendingMessage{},
		mu:      sync.Mutex{},
	}
}

func (s *MessageQueue) Push(m *ProtocolMessage, seqNum uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending = append(s.pending, PendingMessage{
		Message: m,
		SeqNum:  seqNum,
	})
}

func (s *MessageQueue) Messages() []*ProtocolMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending := []*ProtocolMessage{}
	for _, m := range s.pending {
		pending = append(pending, m.Message)
	}
	return pending
}

func (s *MessageQueue) Acknowledge(seqNum uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending := []PendingMessage{}
	for _, m := range s.pending {
		if m.SeqNum > seqNum {
			pending = append(pending, m)
		}
	}
	s.pending = pending
}
