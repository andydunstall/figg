package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlidingWindow_AddOneMessageThenACK(t *testing.T) {
	w := newSlidingWindow(3)

	// Add a message and check returned.
	assert.Equal(t, uint64(0), w.Push("A", []byte("1"), nil))
	assert.Equal(t, []unackedMessage{
		{
			Topic:  "A",
			Data:   []byte("1"),
			SeqNum: 0,
			OnACK:  nil,
		},
	}, w.Messages())

	// Acknowledge the message and check removed.
	w.Acknowledge(0)
	assert.Equal(t, []unackedMessage{}, w.Messages())

	// Add another message and check returned.
	assert.Equal(t, uint64(1), w.Push("B", []byte("2"), nil))
	assert.Equal(t, []unackedMessage{
		{
			Topic:  "B",
			Data:   []byte("2"),
			SeqNum: 1,
			OnACK:  nil,
		},
	}, w.Messages())
}

func TestSlidingWindow_AddTwoMessageThenACK(t *testing.T) {
	// Use a window size of 3 so the indicies wrap around.
	w := newSlidingWindow(3)

	// Add two messages message and check returned.
	assert.Equal(t, uint64(0), w.Push("A", []byte("1"), nil))
	assert.Equal(t, uint64(1), w.Push("B", []byte("2"), nil))
	assert.Equal(t, []unackedMessage{
		{
			Topic:  "A",
			Data:   []byte("1"),
			SeqNum: 0,
			OnACK:  nil,
		},
		{
			Topic:  "B",
			Data:   []byte("2"),
			SeqNum: 1,
			OnACK:  nil,
		},
	}, w.Messages())

	// Acknowledge the first message message and check removed.
	w.Acknowledge(0)
	assert.Equal(t, []unackedMessage{
		{
			Topic:  "B",
			Data:   []byte("2"),
			SeqNum: 1,
			OnACK:  nil,
		},
	}, w.Messages())

	// Add two more messages to fill the buffer and check returned.
	assert.Equal(t, uint64(2), w.Push("C", []byte("3"), nil))
	assert.Equal(t, uint64(3), w.Push("D", []byte("4"), nil))

	assert.Equal(t, []unackedMessage{
		{
			Topic:  "B",
			Data:   []byte("2"),
			SeqNum: 1,
			OnACK:  nil,
		},
		{
			Topic:  "C",
			Data:   []byte("3"),
			SeqNum: 2,
			OnACK:  nil,
		},
		{
			Topic:  "D",
			Data:   []byte("4"),
			SeqNum: 3,
			OnACK:  nil,
		},
	}, w.Messages())

	// Ack all but the last and check returned.
	w.Acknowledge(2)
	assert.Equal(t, []unackedMessage{
		{
			Topic:  "D",
			Data:   []byte("4"),
			SeqNum: 3,
			OnACK:  nil,
		},
	}, w.Messages())
}

func TestSlidingWindow_AckMessages(t *testing.T) {
	w := newSlidingWindow(3)

	// Add a message and check returned.
	firstAcked := false
	w.Push("A", []byte("1"), func() {
		firstAcked = true
	})
	secondAcked := false
	w.Push("B", []byte("2"), func() {
		secondAcked = true
	})
	thirdAcked := false
	w.Push("B", []byte("2"), func() {
		secondAcked = true
	})

	w.Acknowledge(2)

	assert.Equal(t, true, firstAcked)
	assert.Equal(t, true, secondAcked)
	assert.Equal(t, false, thirdAcked)
}
