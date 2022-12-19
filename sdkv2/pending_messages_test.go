package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPendingMessages_AddPending(t *testing.T) {
	pending := newPendingMessages()

	assert.Equal(t, uint64(0), pending.Add("foo", []byte("foo"), nil))
	assert.Equal(t, uint64(1), pending.Add("bar", []byte("bar"), nil))
	assert.Equal(t, uint64(2), pending.Add("car", []byte("car"), nil))

	assert.Equal(t, []pendingMessage{
		{
			Topic:  "foo",
			Data:   []byte("foo"),
			SeqNum: 0,
			OnACK:  nil,
		},
		{
			Topic:  "bar",
			Data:   []byte("bar"),
			SeqNum: 1,
			OnACK:  nil,
		},
		{
			Topic:  "car",
			Data:   []byte("car"),
			SeqNum: 2,
			OnACK:  nil,
		},
	}, pending.Messages())
}

func TestPendingMessages_ACKPending(t *testing.T) {
	pending := newPendingMessages()

	fooAcked := false
	assert.Equal(t, uint64(0), pending.Add("foo", []byte("foo"), func() {
		fooAcked = true
	}))
	barAcked := false
	assert.Equal(t, uint64(1), pending.Add("bar", []byte("bar"), func() {
		barAcked = true
	}))
	carAcked := false
	assert.Equal(t, uint64(2), pending.Add("car", []byte("car"), func() {
		carAcked = true
	}))

	pending.Acknowledge(1)

	assert.True(t, fooAcked)
	assert.True(t, barAcked)
	assert.False(t, carAcked)

	assert.Equal(t, 1, len(pending.Messages()))
	assert.Equal(t, "car", pending.Messages()[0].Topic)
}
