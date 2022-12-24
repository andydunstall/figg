package server

import (
	"testing"

	"github.com/andydunstall/figg/server/pkg/topic"
	"github.com/andydunstall/figg/utils"
	"github.com/stretchr/testify/assert"
)

func TestConnection_Attach(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	fakeConn.Push(utils.EncodeAttachMessage("foo"))

	assert.Nil(t, conn.Recv())
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachedMessage("foo", 0))
}

func TestConnection_AttachFromOffset(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	fakeConn.Push(utils.EncodeAttachFromOffsetMessage("foo", 0xff))

	assert.Nil(t, conn.Recv())
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachedMessage("foo", 0xff))
}

func TestConnection_Publish(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	// Publish a message and expect to be ACK'ed
	for seqNum := uint64(0); seqNum != 10; seqNum++ {
		fakeConn.Push(utils.EncodePublishMessage("foo", seqNum, []byte("bar")))
		assert.Nil(t, conn.Recv())
		assert.Equal(t, fakeConn.NextWritten(), utils.EncodeACKMessage(seqNum))
	}
}

func TestConnection_PublishSendMessagesToAttached(t *testing.T) {
	broker := topic.NewBroker(topic.Options{
		Persisted:   false,
		SegmentSize: 1000,
	})

	// Add a connection subscribing to the topic.
	subConn, subFakeConn := newFakeConnectionWithBroker(broker)
	defer subConn.Close()
	subFakeConn.Push(utils.EncodeAttachMessage("foo"))
	assert.Nil(t, subConn.Recv())
	assert.Equal(t, subFakeConn.NextWritten(), utils.EncodeAttachedMessage("foo", 0))

	// Add another connection and publish to the topic.
	pubConn, pubFakeConn := newFakeConnectionWithBroker(broker)
	defer pubConn.Close()
	pubFakeConn.Push(utils.EncodePublishMessage("foo", 0, []byte("bar")))
	assert.Nil(t, pubConn.Recv())

	// Check the subscriber connection receives the message.
	assert.Equal(t, subFakeConn.NextWritten(), utils.EncodeDataMessage("foo", 7, []byte("bar")))
}

func newFakeConnection() (*Connection, *utils.FakeConn) {
	fakeConn := utils.NewFakeConn()
	conn := NewConnection(fakeConn, topic.NewBroker(topic.Options{
		Persisted:   false,
		SegmentSize: 1000,
	}))
	return conn, fakeConn
}

func newFakeConnectionWithBroker(broker *topic.Broker) (*Connection, *utils.FakeConn) {
	fakeConn := utils.NewFakeConn()
	conn := NewConnection(fakeConn, broker)
	return conn, fakeConn
}
