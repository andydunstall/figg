package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopic_PublishMultipleMessages(t *testing.T) {
	topic, err := NewTopic("foo")
	assert.Nil(t, err)

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	b, offset, ok := topic.GetMessage(0)
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)

	b, offset, ok = topic.GetMessage(offset)
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)

	b, offset, ok = topic.GetMessage(offset)
	assert.True(t, ok)
	assert.Equal(t, []byte("car"), b)

	_, _, ok = topic.GetMessage(offset)
	assert.False(t, ok)

}

func TestTopic_PublishOneMessage(t *testing.T) {
	topic, err := NewTopic("foo")
	assert.Nil(t, err)

	topic.Publish([]byte("foo"))

	b, _, ok := topic.GetMessage(0)
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)
}

func TestTopic_GetInitialMessage(t *testing.T) {
	topic, err := NewTopic("foo")
	assert.Nil(t, err)

	_, _, ok := topic.GetMessage(topic.Offset())
	assert.False(t, ok)
}

func TestTopic_GetInitialOffset(t *testing.T) {
	topic, err := NewTopic("foo")
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), topic.Offset())
}
