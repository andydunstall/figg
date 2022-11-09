package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopic_UpdateOffset(t *testing.T) {
	topic := NewTopic()

	topic.OnMessage([]byte("foo"), "1")
	assert.Equal(t, "1", topic.Offset())
	topic.OnMessage([]byte("bar"), "2")
	assert.Equal(t, "2", topic.Offset())
	topic.OnMessage([]byte("car"), "3")
	assert.Equal(t, "3", topic.Offset())
}

func TestTopic_SubscribeToMessage(t *testing.T) {
	topic := NewTopic()

	sub := NewQueueMessageSubscriber()
	topic.Subscribe(sub)

	topic.OnMessage([]byte("foo"), "1")
	topic.OnMessage([]byte("bar"), "2")

	b, ok := sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)
	b, ok = sub.Next()
	assert.False(t, ok)
}

func TestTopic_Unsubscribe(t *testing.T) {
	topic := NewTopic()

	sub := NewQueueMessageSubscriber()
	topic.Subscribe(sub)
	topic.Unsubscribe(sub)

	topic.OnMessage([]byte("foo"), "1")
	topic.OnMessage([]byte("bar"), "2")

	_, ok := sub.Next()
	assert.False(t, ok)
}

func TestTopic_MultipleSubscribers(t *testing.T) {
	topic := NewTopic()

	sub1 := NewQueueMessageSubscriber()
	assert.True(t, topic.Subscribe(sub1))
	sub2 := NewQueueMessageSubscriber()
	assert.False(t, topic.Subscribe(sub2))

	topic.OnMessage([]byte("foo"), "1")

	b, ok := sub1.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub2.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
}
