package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopics_SubscribeToMessage(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub))

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
	assert.True(t, topics.Subscribe("topic1", sub))
	topics.Unsubscribe("topic1", sub)

	topics.OnMessage("topic1", []byte("foo"))
	topics.OnMessage("topic1", []byte("bar"))

	_, ok := sub.Next()
	assert.False(t, ok)
}

func TestTopics_SubscribeMultipleTopics(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub))
	assert.True(t, topics.Subscribe("topic2", sub))
	assert.True(t, topics.Subscribe("topic3", sub))

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

func TestTopics_SubscribeTopicWithMultipleSubscriptions(t *testing.T) {
	topics := NewTopics()

	sub1 := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub1))
	sub2 := NewQueueMessageSubscriber()
	assert.False(t, topics.Subscribe("topic1", sub2))

	topics.OnMessage("topic1", []byte("foo"))

	b, ok := sub1.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub2.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
}
