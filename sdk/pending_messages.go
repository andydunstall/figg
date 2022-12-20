package figg

import (
	"sync"
)

type pendingMessage struct {
	Topic  string
	Data   []byte
	SeqNum uint64
	OnACK  func()
}

// pendingMessages contains published messages that are waiting to be
// acknowledged.
type pendingMessages struct {
	pending []pendingMessage
	seqNum  uint64
	mu      sync.Mutex
}

func newPendingMessages() *pendingMessages {
	return &pendingMessages{
		pending: []pendingMessage{},
		seqNum:  0,
		mu:      sync.Mutex{},
	}
}

// Add adds a message and returns the messages sequence number.
func (m *pendingMessages) Add(topic string, data []byte, onACK func()) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	seqNum := m.seqNum
	m.seqNum++

	m.pending = append(m.pending, pendingMessage{
		Topic:  topic,
		Data:   data,
		SeqNum: seqNum,
		OnACK:  onACK,
	})
	return seqNum
}

func (m *pendingMessages) Messages() []pendingMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	pending := make([]pendingMessage, 0, len(m.pending))
	for _, p := range m.pending {
		pending = append(pending, p)
	}
	return pending
}

func (m *pendingMessages) Acknowledge(seqNum uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.pending) == 0 {
		return
	}

	firstSeqNum := m.pending[0]
	idx := seqNum - firstSeqNum.SeqNum

	acknowledged := m.pending[:idx+1]
	m.pending = m.pending[idx+1:]

	for _, m := range acknowledged {
		if m.OnACK != nil {
			m.OnACK()
		}
	}
}
