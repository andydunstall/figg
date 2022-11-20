package figg

import (
	"container/list"
	"testing"

	"github.com/stretchr/testify/assert"
)

type messageQueue struct {
	messages *list.List
}

func newMessageQueue() *messageQueue {
	return &messageQueue{
		messages: list.New(),
	}
}

func (q *messageQueue) Next() ([]byte, bool) {
	if q.messages.Len() == 0 {
		return nil, false
	}

	m := q.messages.Front()
	q.messages.Remove(m)
	return m.Value.([]byte), true
}

func (q *messageQueue) Push(m []byte) {
	q.messages.PushBack(m)
}

func TestTopic_UpdateOffset(t *testing.T) {
	topic := NewTopic("mytopic")

	topic.OnMessage([]byte("foo"), "1")
	assert.Equal(t, "1", topic.Offset())
	topic.OnMessage([]byte("bar"), "2")
	assert.Equal(t, "2", topic.Offset())
	topic.OnMessage([]byte("car"), "3")
	assert.Equal(t, "3", topic.Offset())
}

func TestTopic_SubscribeToMessage(t *testing.T) {
	topic := NewTopic("mytopic")

	q := newMessageQueue()
	topic.Subscribe(func(topicName string, m []byte) {
		q.Push(m)
	})

	topic.OnMessage([]byte("foo"), "1")
	topic.OnMessage([]byte("bar"), "2")

	b, ok := q.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = q.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)
	b, ok = q.Next()
	assert.False(t, ok)
}

func TestTopic_Unsubscribe(t *testing.T) {
	topic := NewTopic("mytopic")

	q := newMessageQueue()
	sub, _ := topic.Subscribe(func(topicName string, m []byte) {
		q.Push(m)
	})
	topic.Unsubscribe(sub)

	topic.OnMessage([]byte("foo"), "1")
	topic.OnMessage([]byte("bar"), "2")

	_, ok := q.Next()
	assert.False(t, ok)
}

func TestTopic_MultipleSubscribers(t *testing.T) {
	topic := NewTopic("mytopic")

	q1 := newMessageQueue()
	_, activated := topic.Subscribe(func(topicName string, m []byte) {
		q1.Push(m)
	})
	assert.True(t, activated)
	q2 := newMessageQueue()
	_, activated = topic.Subscribe(func(topicName string, m []byte) {
		q2.Push(m)
	})

	topic.OnMessage([]byte("foo"), "1")

	b, ok := q1.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
	b, ok = q2.Next()
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
}
