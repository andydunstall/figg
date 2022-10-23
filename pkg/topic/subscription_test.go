package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscription_SubscribeLatest(t *testing.T) {
	topic := NewTopic()
	conn := NewFakeConn()
	sub := NewSubscription(topic, conn)
	defer sub.Shutdown()

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	assert.Equal(t, Message{
		Offset:  1,
		Message: []byte("foo"),
	}, <-conn.Sent)
	assert.Equal(t, Message{
		Offset:  2,
		Message: []byte("bar"),
	}, <-conn.Sent)
	assert.Equal(t, Message{
		Offset:  3,
		Message: []byte("car"),
	}, <-conn.Sent)
}

func TestSubscription_SubscribeRecover(t *testing.T) {
	topic := NewTopic()

	// Publish 2 messages prior to subscribing.
	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))

	conn := NewFakeConn()
	sub := NewSubscriptionWithOffset(topic, conn, 0)
	defer sub.Shutdown()

	// Publish 2 messages prior after subscribing.
	topic.Publish([]byte("baz"))
	topic.Publish([]byte("car"))

	assert.Equal(t, Message{
		Offset:  1,
		Message: []byte("foo"),
	}, <-conn.Sent)
	assert.Equal(t, Message{
		Offset:  2,
		Message: []byte("bar"),
	}, <-conn.Sent)
	assert.Equal(t, Message{
		Offset:  3,
		Message: []byte("baz"),
	}, <-conn.Sent)
	assert.Equal(t, Message{
		Offset:  4,
		Message: []byte("car"),
	}, <-conn.Sent)
}
