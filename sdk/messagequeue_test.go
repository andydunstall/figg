package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageQueue_GetPendingMessages(t *testing.T) {
	queue := NewMessageQueue()
	queue.Push([]byte("A"), 0, nil)
	queue.Push([]byte("B"), 1, nil)
	queue.Push([]byte("C"), 2, nil)

	pending := queue.Messages()
	assert.Equal(t, 3, len(pending))
	assert.Equal(t, []byte("A"), pending[0])
	assert.Equal(t, []byte("B"), pending[1])
	assert.Equal(t, []byte("C"), pending[2])
}

func TestMessageQueue_ACKMessages(t *testing.T) {
	queue := NewMessageQueue()
	queue.Push([]byte("A"), 0, nil)
	queue.Push([]byte("B"), 1, nil)
	queue.Push([]byte("C"), 2, nil)
	queue.Push([]byte("D"), 3, nil)

	queue.Acknowledge(2)

	pending := queue.Messages()
	assert.Equal(t, 1, len(pending))
	assert.Equal(t, []byte("D"), pending[0])
}

func TestMessageQueue_ACKMessagesWithCallback(t *testing.T) {
	acknowledged := 0
	cb := func(err error) {
		acknowledged++
	}

	queue := NewMessageQueue()
	queue.Push([]byte("A"), 0, cb)
	queue.Push([]byte("B"), 1, cb)
	queue.Push([]byte("C"), 2, cb)

	queue.Acknowledge(2)

	assert.Equal(t, 3, acknowledged)
}

func BenchmarkMessageQueue_Acknowledge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		queue := NewMessageQueue()
		for i := uint64(0); i != 512; i++ {
			queue.Push([]byte("foo"), i, nil)
		}
		for i := uint64(0); i != 512; i++ {
			queue.Acknowledge(i)
		}
	}
}
