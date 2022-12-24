package topic

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/andydunstall/figg/service/pkg/commitlog"
)

type Attachment interface {
	Send(ctx context.Context, m Message)
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
func NewSubscription(attachment Attachment, topic *Topic) (*Subscription, uint64) {
	// Use the offset of the last message in the topic.
	return NewSubscriptionFromOffset(attachment, topic, topic.Offset())
}

// NewSubscriptionFromOffset creates a subscription to the given topic, starting
// at the next message after the given offset. If the offset is less than the
// earliest message retained by the topic, will subscribe from that earliest
// retained message.
func NewSubscriptionFromOffset(attachment Attachment, topic *Topic, offset uint64) (*Subscription, uint64) {
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
	return s, offset
}

// Notify notifys the subscriber about a new message.
func (s *Subscription) Notify(m Message) {
	s.attachment.Send(nil, m)
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
		m, err := s.topic.GetMessage(s.offset)
		if err == commitlog.ErrNotFound {
			// If we are up to date, register with the topic for the latest
			// messages. Note checking if we are up to date and registering
			// must be atomic to avoid missing messages.
			if s.topic.SubscribeIfLatest(s.offset, s) {
				return
			}
			// If theres been a new message since we last checked just try
			// again.
			continue
		} else if err != nil {
			// TODO(AD) conn closed?
			fmt.Println(err)
			return
		}

		s.offset += commitlog.PrefixSize
		s.offset += uint64(len(m))

		s.attachment.Send(nil, Message{
			Topic:   s.topic.Name(),
			Message: m,
			Offset:  s.offset,
		})
	}
}
