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

	b, offset, err := topic.GetMessage(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, offset, err = topic.GetMessage(offset)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, offset, err = topic.GetMessage(offset)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	_, _, err = topic.GetMessage(offset)
	assert.Equal(t, ErrNotFound, err)
}

func TestTopic_PublishOneMessage(t *testing.T) {
	topic, err := NewTopic("foo")
	assert.Nil(t, err)

	topic.Publish([]byte("foo"))

	b, _, err := topic.GetMessage(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)
}

func TestTopic_GetInitialMessage(t *testing.T) {
	topic, err := NewTopic("foo")
	assert.Nil(t, err)

	_, _, err = topic.GetMessage(topic.Offset())
	assert.Equal(t, ErrNotFound, err)
}
