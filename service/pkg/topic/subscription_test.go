package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscription_SubscribeLatest(t *testing.T) {
	messageCh := make(chan TopicMessage)
	topic := NewTopic("mytopic")
	sub := NewSubscription(messageCh, topic)
	defer sub.Shutdown()

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "1",
		Message: []byte("foo"),
	}, <-messageCh)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "2",
		Message: []byte("bar"),
	}, <-messageCh)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "3",
		Message: []byte("car"),
	}, <-messageCh)
}

func TestSubscription_SubscribeRecover(t *testing.T) {
	topic := NewTopic("mytopic")

	// Publish 2 messages prior to subscribing.
	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))

	messageCh := make(chan TopicMessage)
	sub := NewSubscriptionFromOffset(messageCh, topic, 0)
	defer sub.Shutdown()

	// Publish 2 messages prior after subscribing.
	topic.Publish([]byte("baz"))
	topic.Publish([]byte("car"))

	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "1",
		Message: []byte("foo"),
	}, <-messageCh)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "2",
		Message: []byte("bar"),
	}, <-messageCh)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "3",
		Message: []byte("baz"),
	}, <-messageCh)
	assert.Equal(t, TopicMessage{
		Topic:   "mytopic",
		Offset:  "4",
		Message: []byte("car"),
	}, <-messageCh)
}
