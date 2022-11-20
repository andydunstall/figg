package figg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopics_UpdateOffset(t *testing.T) {
	topics := NewTopics()

	topics.Subscribe("topic1", func(topicName string, m []byte) {})
	topics.Subscribe("topic2", func(topicName string, m []byte) {})
	topics.Subscribe("topic3", func(topicName string, m []byte) {})

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

	q := newMessageQueue()
	topics.Subscribe("topic1", func(topicName string, m []byte) {
		q.Push(m)
	})

	topics.OnMessage("topic1", []byte("foo"), "1")
	topics.OnMessage("topic1", []byte("bar"), "2")

	b, ok := q.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = q.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)
	b, ok = q.Next()
	assert.False(t, ok)
}

func TestTopics_Unsubscribe(t *testing.T) {
	topics := NewTopics()

	q := newMessageQueue()
	sub, _ := topics.Subscribe("topic1", func(topicName string, m []byte) {
		q.Push(m)
	})
	topics.Unsubscribe("topic1", sub)

	topics.OnMessage("topic1", []byte("foo"), "1")
	topics.OnMessage("topic1", []byte("bar"), "2")

	_, ok := q.Next()
	assert.False(t, ok)
}

func TestTopics_ListSubscribedTopics(t *testing.T) {
	topics := NewTopics()

	topics.Subscribe("topic1", func(topicName string, m []byte) {})
	topics.Subscribe("topic2", func(topicName string, m []byte) {})

	names := topics.Topics()
	sort.Strings(names)

	assert.Equal(t, []string{"topic1", "topic2"}, names)
}
