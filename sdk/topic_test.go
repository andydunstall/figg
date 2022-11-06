package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopic_SubscribeToMessage(t *testing.T) {
	topic := NewTopic()

	sub := NewQueueMessageSubscriber()
	topic.Subscribe(sub)

	topic.OnMessage([]byte("foo"))
	topic.OnMessage([]byte("bar"))

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

	topic.OnMessage([]byte("foo"))
	topic.OnMessage([]byte("bar"))

	_, ok := sub.Next()
	assert.False(t, ok)
}

func TestTopic_MultipleSubscribers(t *testing.T) {
	topic := NewTopic()

	sub1 := NewQueueMessageSubscriber()
	topic.Subscribe(sub1)
	sub2 := NewQueueMessageSubscriber()
	topic.Subscribe(sub2)

	topic.OnMessage([]byte("foo"))

	b, ok := sub1.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub2.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
}
