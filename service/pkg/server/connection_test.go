package server

import (
	"testing"

	"github.com/andydunstall/figg/service/pkg/topic"
	"github.com/andydunstall/figg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeConn struct {
	Incoming []byte
	Outgoing [][]byte
}

func (c *fakeConn) Read(b []byte) (n int, err error) {
	end := len(c.Incoming)
	if end > len(b) {
		end = len(b)
	}

	for i := 0; i < end; i++ {
		b[i] = c.Incoming[i]
	}
	return end, nil
}

func (c *fakeConn) Write(b []byte) (n int, err error) {
	if c.Outgoing == nil {
		c.Outgoing = [][]byte{}
	}

	c.Outgoing = append(c.Outgoing, b)
	return 0, nil
}

func (c *fakeConn) NextOutgoing() []byte {
	if len(c.Outgoing) == 0 {
		return nil
	} else {
		next := c.Outgoing[0]
		c.Outgoing = c.Outgoing[1:]
		return next
	}
}

func (c *fakeConn) Close() error {
	return nil
}

func TestConnection_Attach(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := NewConnection(fakeConn, topic.NewBroker("data/"+uuid.New().String()))

	fakeConn.Incoming = utils.EncodeAttachMessage("foo")

	assert.Nil(t, conn.Recv())
	assert.Equal(t, fakeConn.NextOutgoing(), utils.EncodeAttachedMessage("foo", 0))
}

// TODO(AD) attach from offset

// TODO(AD) Test attached returns offset

func TestConnection_Publish(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := NewConnection(fakeConn, topic.NewBroker("data/"+uuid.New().String()))

	// Publish a message and expect to be ACK'ed
	for seqNum := uint64(0); seqNum != 10; seqNum++ {
		fakeConn.Incoming = utils.EncodePublishMessage("foo", seqNum, []byte("bar"))
		assert.Nil(t, conn.Recv())
		assert.Equal(t, fakeConn.NextOutgoing(), utils.EncodeACKMessage(seqNum))
	}
}
