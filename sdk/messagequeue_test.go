package figg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func fakePublishMessage(seqNum uint64) *ProtocolMessage {
	return &ProtocolMessage{Publish: &PublishMessage{SeqNum: seqNum}}
}

func TestMessageQueue_GetPendingMessages(t *testing.T) {
	queue := NewMessageQueue()
	queue.Push(fakePublishMessage(0), 0, nil)
	queue.Push(fakePublishMessage(1), 1, nil)
	queue.Push(fakePublishMessage(2), 2, nil)

	pending := queue.Messages()
	assert.Equal(t, 3, len(pending))
	assert.Equal(t, uint64(0), pending[0].Publish.SeqNum)
	assert.Equal(t, uint64(1), pending[1].Publish.SeqNum)
	assert.Equal(t, uint64(2), pending[2].Publish.SeqNum)
}

func TestMessageQueue_ACKMessages(t *testing.T) {
	queue := NewMessageQueue()
	queue.Push(fakePublishMessage(0), 0, nil)
	queue.Push(fakePublishMessage(1), 1, nil)
	queue.Push(fakePublishMessage(2), 2, nil)
	queue.Push(fakePublishMessage(3), 3, nil)

	queue.Acknowledge(2)

	pending := queue.Messages()
	assert.Equal(t, 1, len(pending))
	assert.Equal(t, uint64(3), pending[0].Publish.SeqNum)
}
