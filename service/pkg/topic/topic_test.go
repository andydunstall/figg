package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopic_PublishMultipleMessages(t *testing.T) {
	topic := NewTopic()

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	assert.Equal(t, uint64(3), topic.Offset())
	b, offset, ok := topic.GetMessage(topic.Offset())
	assert.Equal(t, string(b), "car")
	assert.Equal(t, uint64(3), offset)
	assert.True(t, ok)
}

func TestTopic_PublishOneMessage(t *testing.T) {
	topic := NewTopic()

	topic.Publish([]byte("foo"))

	assert.Equal(t, uint64(1), topic.Offset())
	b, offset, ok := topic.GetMessage(topic.Offset())
	assert.Equal(t, string(b), "foo")
	assert.Equal(t, uint64(1), offset)
	assert.True(t, ok)
}

func TestTopic_GetInitialMessage(t *testing.T) {
	topic := NewTopic()

	_, _, ok := topic.GetMessage(topic.Offset())
	assert.False(t, ok)
}

func TestTopic_GetInitialOffset(t *testing.T) {
	topic := NewTopic()
	assert.Equal(t, uint64(0), topic.Offset())
}
