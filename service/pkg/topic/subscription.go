package topic

import (
	"sync"
	"sync/atomic"
)

type TopicMessage struct {
	Topic   string
	Message []byte
	Offset  uint64
}

// Subscription reads messages from the topic and sends to the connection.
type Subscription struct {
	topic *Topic
	// lastOffset is the offset of the last processed message in the topic.
	lastOffset uint64

	messageCh chan<- TopicMessage

	cv       *sync.Cond
	mu       *sync.Mutex
	shutdown int32
}

// NewSubscription creates a subscription to the given topic starting from the
// next message in the topic.
func NewSubscription(messageCh chan<- TopicMessage, topic *Topic) *Subscription {
	// Use the offset of the last message in the topic.
	return NewSubscriptionFromOffset(messageCh, topic, topic.Offset())
}

// NewSubscriptionFromOffset creates a subscription to the given topic, starting
// at the next message after the given offset. If the offset is less than the
// earliest message retained by the topic, will subscribe from that earliest
// retained message.
func NewSubscriptionFromOffset(messageCh chan<- TopicMessage, topic *Topic, lastOffset uint64) *Subscription {
	mu := &sync.Mutex{}
	s := &Subscription{
		topic:      topic,
		lastOffset: lastOffset,
		messageCh:  messageCh,
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

			s.messageCh <- TopicMessage{
				Topic:   s.topic.Name(),
				Message: m,
				Offset:  offset,
			}
			s.lastOffset = offset
		}

		// Block until we are either shut down or there is a new message on
		// the topic.
		s.mu.Lock()
		s.cv.Wait()
		s.mu.Unlock()
	}
}