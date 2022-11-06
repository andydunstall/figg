package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopics_SubscribeToMessage(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	topics.Subscribe("topic1", sub)

	topics.OnMessage("topic1", []byte("foo"))
	topics.OnMessage("topic1", []byte("bar"))

	b, ok := sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)
	b, ok = sub.Next()
	assert.False(t, ok)
}

func TestTopics_Unsubscribe(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	topics.Subscribe("topic1", sub)
	topics.Unsubscribe("topic1", sub)

	topics.OnMessage("topic1", []byte("foo"))
	topics.OnMessage("topic1", []byte("bar"))

	_, ok := sub.Next()
	assert.False(t, ok)
}

func TestTopics_SubscribeMultipleTopics(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	topics.Subscribe("topic1", sub)
	topics.Subscribe("topic2", sub)
	topics.Subscribe("topic3", sub)

	topics.OnMessage("topic1", []byte("foo"))
	topics.OnMessage("topic2", []byte("bar"))
	topics.OnMessage("topic3", []byte("car"))

	b, ok := sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)
	b, ok = sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("car"), b)
	b, ok = sub.Next()
	assert.False(t, ok)
}
