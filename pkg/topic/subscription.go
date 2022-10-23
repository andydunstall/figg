package topic

import (
	"sync"
	"sync/atomic"
)

type Message struct {
	Offset  uint64
	Message []byte
}

type Conn interface {
	Send(offset uint64, m []byte) error
	Recv() ([]byte, error)
}

// Subscription reads messages from the topic and sends to the connection.
type Subscription struct {
	topic *Topic
	conn  Conn
	// lastOffset is the offset of the last processed message in the topic.
	lastOffset uint64

	cv       *sync.Cond
	mu       *sync.Mutex
	shutdown int32
}

// NewSubscription creates a subscription to the given topic starting from the
// next message in the topic.
func NewSubscription(topic *Topic, conn Conn) *Subscription {
	// Use the offset of the last message in the topic.
	return NewSubscriptionWithOffset(topic, conn, topic.Offset())
}

// NewSubscriptionWithOffset creates a subscription to the given topic, starting
// at the next message after the given offset. If the offset is less than the
// earliest message retained by the topic, will subscribe from that earliest
// retained message.
func NewSubscriptionWithOffset(topic *Topic, conn Conn, lastOffset uint64) *Subscription {
	mu := &sync.Mutex{}
	s := &Subscription{
		topic:      topic,
		conn:       conn,
		lastOffset: lastOffset,
		cv:         sync.NewCond(mu),
		mu:         mu,
	}
	topic.Subscribe(s)
	go s.sendLoop()
	return s
}

// Signal signals the send loop to check for new messages on the topic.
func (s *Subscription) Signal() {
	s.mu.Lock()
	s.cv.Signal()
	s.mu.Unlock()
}

// Shutdown unsubscribes and stops the send loop.
func (s *Subscription) Shutdown() {
	s.topic.Unsubscribe(s)

	// Notify the send loop to stop (must signal it to wake up to check the
	// shutdown flag).
	atomic.StoreInt32(&s.shutdown, 1)
	s.mu.Lock()
	s.cv.Signal()
	s.mu.Unlock()
}

func (s *Subscription) sendLoop() {
	for {
		if s := atomic.LoadInt32(&s.shutdown); s != 0 {
			return
		}

		// Note doesn't need to be locked by mu as only the sendLoop updates
		// s.lastOffset.
		for s.lastOffset < s.topic.Offset() {
			// Note if there is no message with offset s.lastOffset+1, will
			// round up to the earliest message on the topic.
			m, offset, ok := s.topic.GetMessage(s.lastOffset + 1)
			if !ok {
				break
			}

			// Only update the offset once its been sent to the subscriber. If
			// the connection closes expect the read loop to close the
			// subscriber and the client can resume from the last offset.
			if err := s.conn.Send(offset, m); err != nil {
				break
			} else {
				s.lastOffset = offset
			}
		}

		// Block until we are either shut down or there is a new message on
		// the topic.
		s.mu.Lock()
		s.cv.Wait()
		s.mu.Unlock()
	}
}
