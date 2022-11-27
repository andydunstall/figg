package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeAttachment struct {
	Ch chan TopicMessage
}

func newFakeAttachment() *fakeAttachment {
	return &fakeAttachment{
		Ch: make(chan TopicMessage),
	}
}

func (a *fakeAttachment) Send(m TopicMessage) {
	a.Ch <- m
}

func TestSubscription_SubscribeLatest(t *testing.T) {
	topic := NewTopic("mytopic")
	attachment := newFakeAttachment()
	sub := NewSubscription(attachment, topic)
	defer sub.Shutdown()

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "1",
		Message: []byte("foo"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "2",
		Message: []byte("bar"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "3",
		Message: []byte("car"),
	}, <-attachment.Ch)
}

func TestSubscription_SubscribeRecover(t *testing.T) {
	topic := NewTopic("mytopic")

	// Publish 2 messages prior to subscribing.
	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))

	attachment := newFakeAttachment()
	sub := NewSubscriptionFromOffset(attachment, topic, 0)
	defer sub.Shutdown()

	// Publish 2 messages prior after subscribing.
	topic.Publish([]byte("baz"))
	topic.Publish([]byte("car"))

	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "1",
		Message: []byte("foo"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "2",
		Message: []byte("bar"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "3",
		Message: []byte("baz"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "4",
		Message: []byte("car"),
	}, <-attachment.Ch)
}
