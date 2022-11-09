package figg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopics_UpdateOffset(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub))
	assert.True(t, topics.Subscribe("topic2", sub))
	assert.True(t, topics.Subscribe("topic3", sub))

	topics.OnMessage("topic1", []byte("foo"), "1")
	topics.OnMessage("topic1", []byte("bar"), "2")
	topics.OnMessage("topic2", []byte("bar"), "1")
	topics.OnMessage("topic3", []byte("bar"), "1")
	topics.OnMessage("topic1", []byte("bar"), "3")

	assert.Equal(t, "3", topics.Offset("topic1"))
	assert.Equal(t, "1", topics.Offset("topic2"))
	assert.Equal(t, "1", topics.Offset("topic3"))
}

func TestTopics_SubscribeToMessage(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub))

	topics.OnMessage("topic1", []byte("foo"), "1")
	topics.OnMessage("topic1", []byte("bar"), "2")

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

	topics.OnMessage("topic1", []byte("foo"), "1")
	topics.OnMessage("topic1", []byte("bar"), "2")

	_, ok := sub.Next()
	assert.False(t, ok)
}

func TestTopics_SubscribeMultipleTopics(t *testing.T) {
	topics := NewTopics()

	sub := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub))
	assert.True(t, topics.Subscribe("topic2", sub))
	assert.True(t, topics.Subscribe("topic3", sub))

	topics.OnMessage("topic1", []byte("foo"), "1")
	topics.OnMessage("topic2", []byte("bar"), "1")
	topics.OnMessage("topic3", []byte("car"), "1")

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

	topics.OnMessage("topic1", []byte("foo"), "1")

	b, ok := sub1.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = sub2.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
}

func TestTopics_ListSubscribedTopics(t *testing.T) {
	topics := NewTopics()

	sub1 := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic1", sub1))
	sub2 := NewQueueMessageSubscriber()
	assert.True(t, topics.Subscribe("topic2", sub2))

	names := topics.Topics()
	sort.Strings(names)

	assert.Equal(t, []string{"topic1", "topic2"}, names)
}
