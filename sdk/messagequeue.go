package figg

import (
	"sync"
)

type PendingMessage struct {
	Message []byte
	SeqNum  uint64
	CB      func(err error)
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

// Push adds a message to the queue. Note seqNum MUST be contiguous.
func (s *MessageQueue) Push(m []byte, seqNum uint64, cb func(err error)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pending) != 0 && s.pending[len(s.pending)-1].SeqNum+1 != seqNum {
		panic("non contiguous sequence number")
	}

	s.pending = append(s.pending, PendingMessage{
		Message: m,
		SeqNum:  seqNum,
		CB:      cb,
	})
}

func (s *MessageQueue) Messages() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending := [][]byte{}
	for _, m := range s.pending {
		pending = append(pending, m.Message)
	}
	return pending
}

func (s *MessageQueue) Acknowledge(seqNum uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pending) == 0 {
		return
	}

	firstSeqNum := s.pending[0]
	idx := seqNum - firstSeqNum.SeqNum

	acknowledged := s.pending[:idx+1]
	s.pending = s.pending[idx+1:]

	for _, m := range acknowledged {
		if m.CB != nil {
			m.CB(nil)
		}
	}
}
