package wombat

import (
	"sync"
)

type PendingMessages struct {
	pending []*ProtocolMessage
	mu      sync.Mutex
}

func NewPendingMessages() *PendingMessages {
	return &PendingMessages{
		pending: []*ProtocolMessage{},
		mu:      sync.Mutex{},
	}
}

func (s *PendingMessages) Push(m *ProtocolMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending = append(s.pending, m)
}

func (s *PendingMessages) Get() []*ProtocolMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending := []*ProtocolMessage{}
	for _, m := range s.pending {
		pending = append(pending, m)
	}
	return pending
}
