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
	// offset is the offset of the next message to fetch in the topic.
	// This is only set if the subscriber is resuming.
	offset uint64

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
func NewSubscriptionFromOffset(attachment Attachment, topic *Topic, offset uint64) *Subscription {
	s := &Subscription{
		topic:      topic,
		offset:     offset,
		attachment: attachment,
	}
	// If we are not up to date with the topic run resume to send the backlog.
	if offset == topic.Offset() {
		topic.Subscribe(s)
	} else {
		go s.resumeLoop()
	}
	return s
}

// Notify notifys the subscriber about a new message.
func (s *Subscription) Notify(name string, serial string, m []byte) {
	s.attachment.Send(TopicMessage{
		Topic:   name,
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

// resumeLoop iterates though the topics history until the subscriber is up
// to date, then registers for new messages.
func (s *Subscription) resumeLoop() {
	for {
		if s := atomic.LoadInt32(&s.shutdown); s != 0 {
			return
		}

		// Note if there is no message with offset, will round up to the
		// earliest message on the topic.
		m, offset, err := s.topic.GetMessage(s.offset)
		if err == ErrNotFound {
			// If we are up to date, register with the topic for the latest
			// messages. Note checking if we are up to date and registering
			// must be atomic to avoid missing messages.
			if s.topic.SubscribeIfLatest(offset, s) {
				return
			}
			// If theres been a new message since we last checked just try
			// again.
			continue
		} else if err != nil {
			// TODO(AD)
			panic(err)
		}

		s.attachment.Send(TopicMessage{
			Topic:   s.topic.Name(),
			Message: m,
			Offset:  strconv.FormatUint(offset, 10),
		})
		s.offset = offset
	}
}
