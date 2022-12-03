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
		Ch: make(chan TopicMessage, 64),
	}
}

func (a *fakeAttachment) Send(m TopicMessage) {
	a.Ch <- m
}

func TestSubscription_SubscribeLatest(t *testing.T) {
	topic, err := NewTopic("mytopic")
	assert.Nil(t, err)
	attachment := newFakeAttachment()
	sub := NewSubscription(attachment, topic)
	defer sub.Shutdown()

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "11",
		Message: []byte("foo"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "22",
		Message: []byte("bar"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "33",
		Message: []byte("car"),
	}, <-attachment.Ch)
}

func TestSubscription_SubscribeRecover(t *testing.T) {
	topic, err := NewTopic("mytopic")
	assert.Nil(t, err)

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
		Offset:  "11",
		Message: []byte("foo"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "22",
		Message: []byte("bar"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "33",
		Message: []byte("baz"),
	}, <-attachment.Ch)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "44",
		Message: []byte("car"),
	}, <-attachment.Ch)
}
