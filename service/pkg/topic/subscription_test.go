package topic

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeAttachment struct {
	Ch chan Message
}

func newFakeAttachment() *fakeAttachment {
	return &fakeAttachment{
		Ch: make(chan Message, 64),
	}
}

func (a *fakeAttachment) Send(m Message) {
	a.Ch <- m
}

func TestSubscription_SubscribeLatest(t *testing.T) {
	topic, err := NewTopic("mytopic", "data/"+uuid.New().String())
	assert.Nil(t, err)
	attachment := newFakeAttachment()
	sub := NewSubscription(attachment, topic)
	defer sub.Shutdown()

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "7",
		Message: []byte("foo"),
	}, <-attachment.Ch)
	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "14",
		Message: []byte("bar"),
	}, <-attachment.Ch)
	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "21",
		Message: []byte("car"),
	}, <-attachment.Ch)
}

func TestSubscription_SubscribeRecover(t *testing.T) {
	topic, err := NewTopic("mytopic", "data/"+uuid.New().String())
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

	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "7",
		Message: []byte("foo"),
	}, <-attachment.Ch)
	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "14",
		Message: []byte("bar"),
	}, <-attachment.Ch)
	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "21",
		Message: []byte("baz"),
	}, <-attachment.Ch)
	assert.Equal(t, Message{
		Topic:   "mytopic",
		Offset:  "28",
		Message: []byte("car"),
	}, <-attachment.Ch)
}
