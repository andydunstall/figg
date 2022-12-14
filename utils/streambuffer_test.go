package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamBuffer_PushMessage(t *testing.T) {
	buf := NewStreamBuffer()
	buf.Push([]byte{0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3})

	messageType, m, ok := buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, TypePublish, messageType)
	assert.Equal(t, []byte{0x1, 0x2, 0x3}, m)

	_, _, ok = buf.Next()
	assert.Equal(t, false, ok)
}

func TestStreamBuffer_PushMultiMessage(t *testing.T) {
	buf := NewStreamBuffer()
	buf.Push([]byte{
		0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3,
		0x0, 0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x2, 0x2, 0x3,
		0x0, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x3, 0x2, 0x3,
	})

	messageType, m, ok := buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, TypePublish, messageType)
	assert.Equal(t, []byte{0x1, 0x2, 0x3}, m)

	messageType, m, ok = buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, TypeACK, messageType)
	assert.Equal(t, []byte{0x2, 0x2, 0x3}, m)

	messageType, m, ok = buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, TypePayload, messageType)
	assert.Equal(t, []byte{0x3, 0x2, 0x3}, m)

	_, _, ok = buf.Next()
	assert.Equal(t, false, ok)
}

func TestStreamBuffer_PushMessagePartial(t *testing.T) {
	buf := NewStreamBuffer()
	// Push 1 byte at a time.
	encoded := []byte{0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3}
	for _, b := range encoded {
		buf.Push([]byte{b})
	}

	messageType, m, ok := buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, TypePublish, messageType)
	assert.Equal(t, []byte{0x1, 0x2, 0x3}, m)

	_, _, ok = buf.Next()
	assert.Equal(t, false, ok)
}

func TestStreamBuffer_Empty(t *testing.T) {
	buf := NewStreamBuffer()
	_, _, ok := buf.Next()
	assert.Equal(t, false, ok)
}
