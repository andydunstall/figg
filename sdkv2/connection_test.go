package figg

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeConn struct {
	Incoming [][]byte
	Outgoing []byte
}

func (c *fakeConn) Read(b []byte) (n int, err error) {
	end := len(c.Outgoing)
	if end > len(b) {
		end = len(b)
	}

	for i := 0; i < end; i++ {
		b[i] = c.Outgoing[i]
	}
	return end, nil
}

func (c *fakeConn) Write(b []byte) (n int, err error) {
	c.Incoming = append(c.Incoming, b)
	return 0, nil
}

func (c *fakeConn) NextIncoming() []byte {
	if len(c.Incoming) == 0 {
		return nil
	} else {
		next := c.Incoming[0]
		c.Incoming = c.Incoming[1:]
		return next
	}
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) LocalAddr() net.Addr {
	return nil
}

func (c *fakeConn) RemoteAddr() net.Addr {
	return nil
}

func (c *fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type fakeDialer struct {
	conn net.Conn
}

func (d *fakeDialer) Dial(network string, address string) (net.Conn, error) {
	return d.conn, nil
}

func TestConnection_Attach(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	attached := false
	conn.Attach("foo", func() {
		attached = true
	}, func(m Message) {})

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))
	fakeConn.Outgoing = encodeAttachedMessage("foo", 10)

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_AttachFromOffset(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	attached := false
	conn.AttachFromOffset("foo", 0xff, func() {
		attached = true
	}, func(m Message) {})

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachFromOffsetMessage("foo", 0xff))
	fakeConn.Outgoing = encodeAttachedMessage("foo", 0xff)

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

// Tests when the connection reconnects it resends ATTACH for all pending
// attachment.
func TestConnection_ReattachPendingAttachmentOnReconnect(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	attached := false
	conn.Attach("foo", func() {
		attached = true
	}, func(m Message) {})

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))

	// Reconnect before responding. This should cause the client to resend
	// the ATTACH message.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))
	fakeConn.Outgoing = encodeAttachedMessage("foo", 0xff)

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_ReattachPendingAttachmentFromOffsetOnReconnect(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	attached := false
	conn.AttachFromOffset("foo", 0xff, func() {
		attached = true
	}, func(m Message) {})

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachFromOffsetMessage("foo", 0xff))

	// Reconnect before responding. This should cause the client to resend
	// the ATTACH message.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachFromOffsetMessage("foo", 0xff))
	fakeConn.Outgoing = encodeAttachedMessage("foo", 0xff)

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_ReattachActiveAttachmentOnReconnect(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	attached := false
	conn.Attach("foo", func() {
		attached = true
	}, func(m Message) {})

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))

	// Response with ATTACHED.
	fakeConn.Outgoing = encodeAttachedMessage("foo", 0xff)
	assert.Nil(t, conn.Recv())
	assert.True(t, attached)

	// Reconnect and expect all active topics to be reattached from the returned
	// offset.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachFromOffsetMessage("foo", 0xff))
	fakeConn.Outgoing = encodeAttachedMessage("foo", 0xff)

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_Detach(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	conn.Attach("foo", func() {}, func(m Message) {})
	conn.Detach("foo")

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))
	assert.Equal(t, fakeConn.NextIncoming(), encodeDetachMessage("foo"))
	fakeConn.Outgoing = encodeDetachedMessage("foo")

	assert.Nil(t, conn.Recv())
	assert.Equal(t, 0, len(conn.attachments.Detaching()))
}

func TestConnection_DetachNotAttachedOrAttaching(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	// Detaching a topic thats not attached or attaching should do nothing.
	conn.Detach("foo")
	assert.True(t, fakeConn.NextIncoming() == nil)
}

func TestConnection_ResendDetachingOnReconnect(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	conn.Attach("foo", func() {}, func(m Message) {})
	conn.Detach("foo")

	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))
	assert.Equal(t, fakeConn.NextIncoming(), encodeDetachMessage("foo"))

	// Reconnect before responding. This should cause the client to resend
	// the DETACH message.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextIncoming(), encodeDetachMessage("foo"))

	// Not respond and check clears.
	fakeConn.Outgoing = encodeDetachedMessage("foo")
	assert.Nil(t, conn.Recv())
	assert.Equal(t, 0, len(conn.attachments.Detaching()))

	// Reconnect again and not the client shouldn't do anything.
	conn.Reconnect()
	assert.True(t, fakeConn.NextIncoming() == nil)
}

func TestConnection_Publish(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	conn.Publish("foo", []byte("A"), func() {})
	conn.Publish("foo", []byte("B"), func() {})
	conn.Publish("bar", []byte("C"), func() {})

	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("foo", 0, []byte("A")))
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("foo", 1, []byte("B")))
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("bar", 2, []byte("C")))
}

func TestConnection_PublishRetryOnReconnect(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	conn.Publish("foo", []byte("A"), func() {})
	conn.Publish("foo", []byte("B"), func() {})
	conn.Publish("bar", []byte("C"), func() {})

	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("foo", 0, []byte("A")))
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("foo", 1, []byte("B")))
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("bar", 2, []byte("C")))

	// Reconnect before ACK'ing. Expect to receive the messages again.
	conn.Reconnect()
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("foo", 0, []byte("A")))
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("foo", 1, []byte("B")))
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("bar", 2, []byte("C")))

	// ACK the first 2 messages only.
	fakeConn.Outgoing = encodeACKMessage(1)
	assert.Nil(t, conn.Recv())

	// Reconnect again and now should only get the only unACK'ed message resent.
	conn.Reconnect()
	assert.Equal(t, fakeConn.NextIncoming(), encodePublishMessage("bar", 2, []byte("C")))

	// ACK the final message. Now when reconnecting no publishes should be
	// retried.
	fakeConn.Outgoing = encodeACKMessage(2)
	assert.Nil(t, conn.Recv())
	conn.Reconnect()
	assert.True(t, fakeConn.NextIncoming() == nil)
}

func TestConnection_OnMessage(t *testing.T) {
	fakeConn := &fakeConn{}
	conn := newFakeConnection(fakeConn)

	messages := []Message{}

	// Add attachment.
	conn.Attach("foo", func() {}, func(m Message) {
		data := make([]byte, 0, len(m.Data))
		for _, b := range m.Data {
			data = append(data, b)
		}

		messages = append(messages, Message{
			Data:   data,
			Offset: m.Offset,
		})
	})
	assert.Equal(t, fakeConn.NextIncoming(), encodeAttachMessage("foo"))
	fakeConn.Outgoing = encodeAttachedMessage("foo", 0xff)
	assert.Nil(t, conn.Recv())

	fakeConn.Outgoing = encodeDataMessage("foo", 0x105, []byte("A"))
	assert.Nil(t, conn.Recv())
	fakeConn.Outgoing = encodeDataMessage("foo", 0x110, []byte("B"))
	assert.Nil(t, conn.Recv())
	// Another topic message should be ignored.
	fakeConn.Outgoing = encodeDataMessage("bar", 0x102, []byte("C"))
	assert.Nil(t, conn.Recv())
	fakeConn.Outgoing = encodeDataMessage("foo", 0x115, []byte("D"))
	assert.Nil(t, conn.Recv())

	assert.Equal(t, []Message{
		{
			Data:   []byte("A"),
			Offset: 0x105,
		},
		{
			Data:   []byte("B"),
			Offset: 0x110,
		},
		{
			Data:   []byte("D"),
			Offset: 0x115,
		},
	}, messages)
}

func newFakeConnection(fakeConn *fakeConn) *connection {
	dialer := &fakeDialer{
		conn: fakeConn,
	}
	opts := defaultOptions("1.2.3.4:123")
	opts.Dialer = dialer

	conn := newConnection(nil, opts)
	conn.Connect()
	return conn
}
