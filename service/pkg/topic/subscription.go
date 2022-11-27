package topic

import (
	"strconv"
	"sync/atomic"
)

type Attachment interface {
	Send(m TopicMessage)
}

type TopicMessage struct {
	Topic   string
	Message []byte
	Offset  string
}

// Subscription reads messages from the topic and sends to the connection.
type Subscription struct {
	topic *Topic
	// lastOffset is the offset of the last processed message in the topic.
	// This is only set if the subscriber is resuming.
	lastOffset uint64

	attachment Attachment

	shutdown int32
}

// NewSubscription creates a subscription to the given topic starting from the
// next message in the topic.
func NewSubscription(attachment Attachment, topic *Topic) *Subscription {
	// Use the offset of the last message in the topic.
	return NewSubscriptionFromOffset(attachment, topic, topic.Offset())
}

// NewSubscriptionFromOffset creates a subscription to the given topic, starting
// at the next message after the given offset. If the offset is less than the
// earliest message retained by the topic, will subscribe from that earliest
// retained message.
func NewSubscriptionFromOffset(attachment Attachment, topic *Topic, lastOffset uint64) *Subscription {
	s := &Subscription{
		topic:      topic,
		lastOffset: lastOffset,
		attachment: attachment,
	}
	go s.sendLoop()
	return s
}

// Notify notifys the subscriber about a new message.
func (s *Subscription) Notify(serial string, m []byte) {
	s.attachment.Send(TopicMessage{
		Topic:   s.topic.Name(),
		Message: m,
		Offset:  serial,
	})
}

// Shutdown unsubscribes and stops the send loop.
func (s *Subscription) Shutdown() {
	s.topic.Unsubscribe(s)

	// Notify the send loop to stop (must signal it to wake up to check the
	// shutdown flag).
	atomic.StoreInt32(&s.shutdown, 1)
}

func (s *Subscription) sendLoop() {
	for {
		if s := atomic.LoadInt32(&s.shutdown); s != 0 {
			return
		}

		// Note if there is no message with offset s.lastOffset+1, will
		// round up to the earliest message on the topic.
		m, offset, ok := s.topic.GetMessage(s.lastOffset + 1)
		if !ok {
			// If we are up to date, register with the topic for the latest
			// messages. Note checking if we are up to date and registering
			// must be atomic to avoid missing messages.
			if s.topic.SubscribeIfLatest(s.lastOffset, s) {
				return
			}
			// If theres been a new message since we last checked just try
			// again.
			continue
		}

		s.attachment.Send(TopicMessage{
			Topic:   s.topic.Name(),
			Message: m,
			Offset:  strconv.FormatUint(offset, 10),
		})
		s.lastOffset = offset
	}
}
