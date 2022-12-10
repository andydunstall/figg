package conn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamBuffer_PushMessage(t *testing.T) {
	buf := NewStreamBuffer()
	buf.Push([]byte{0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3})

	m, ok := buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{0x1, 0x2, 0x3}, m)

	_, ok = buf.Next()
	assert.Equal(t, false, ok)
}

func TestStreamBuffer_PushMultiMessage(t *testing.T) {
	buf := NewStreamBuffer()
	buf.Push([]byte{
		0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3,
		0x0, 0x0, 0x0, 0x3, 0x2, 0x2, 0x3,
		0x0, 0x0, 0x0, 0x3, 0x3, 0x2, 0x3,
	})

	m, ok := buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{0x1, 0x2, 0x3}, m)

	m, ok = buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{0x2, 0x2, 0x3}, m)

	m, ok = buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{0x3, 0x2, 0x3}, m)

	_, ok = buf.Next()
	assert.Equal(t, false, ok)
}

func TestStreamBuffer_PushMessagePartial(t *testing.T) {
	buf := NewStreamBuffer()
	// Push 1 byte at a time.
	encoded := []byte{0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3}
	for _, b := range encoded {
		buf.Push([]byte{b})
	}

	m, ok := buf.Next()
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{0x1, 0x2, 0x3}, m)

	_, ok = buf.Next()
	assert.Equal(t, false, ok)
}

func TestStreamBuffer_Empty(t *testing.T) {
	buf := NewStreamBuffer()
	_, ok := buf.Next()
	assert.Equal(t, false, ok)
}
