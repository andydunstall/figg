package figg

import (
	"net"
	"testing"

	"github.com/andydunstall/figg/utils"
	"github.com/stretchr/testify/assert"
)

func TestConnection_Attach(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	attached := false
	conn.Attach("foo", func() {
		attached = true
	}, func(m *Message) {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 10))

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_AttachFromOffset(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	attached := false
	conn.AttachFromOffset("foo", 0xff, func() {
		attached = true
	}, func(m *Message) {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachFromOffsetMessage("foo", 0xff))
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 0xff))

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

// Tests when the connection reconnects it resends ATTACH for all pending
// attachment.
func TestConnection_ReattachPendingAttachmentOnReconnect(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	attached := false
	conn.Attach("foo", func() {
		attached = true
	}, func(m *Message) {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))

	// Reconnect before responding. This should cause the client to resend
	// the ATTACH message.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 0xff))

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_ReattachPendingAttachmentFromOffsetOnReconnect(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	attached := false
	conn.AttachFromOffset("foo", 0xff, func() {
		attached = true
	}, func(m *Message) {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachFromOffsetMessage("foo", 0xff))

	// Reconnect before responding. This should cause the client to resend
	// the ATTACH message.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachFromOffsetMessage("foo", 0xff))
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 0xff))

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_ReattachActiveAttachmentOnReconnect(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	attached := false
	conn.Attach("foo", func() {
		attached = true
	}, func(m *Message) {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))

	// Response with ATTACHED.
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 0xff))
	assert.Nil(t, conn.Recv())
	assert.True(t, attached)

	// Reconnect and expect all active topics to be reattached from the returned
	// offset.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachFromOffsetMessage("foo", 0xff))
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 0xff))

	assert.Nil(t, conn.Recv())
	assert.True(t, attached)
}

func TestConnection_Detach(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	conn.Attach("foo", func() {}, func(m *Message) {})
	conn.Detach("foo")

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeDetachMessage("foo"))
	fakeConn.Push(utils.EncodeDetachedMessage("foo"))

	assert.Nil(t, conn.Recv())
	assert.Equal(t, 0, len(conn.attachments.Detaching()))
}

func TestConnection_ResendDetachingOnReconnect(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	conn.Attach("foo", func() {}, func(m *Message) {})
	conn.Detach("foo")

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeDetachMessage("foo"))

	// Reconnect before responding. This should cause the client to resend
	// the DETACH message.
	conn.Reconnect()

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeDetachMessage("foo"))

	// Not respond and check clears.
	fakeConn.Push(utils.EncodeDetachedMessage("foo"))
	assert.Nil(t, conn.Recv())
	assert.Equal(t, 0, len(conn.attachments.Detaching()))
}

func TestConnection_Publish(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	conn.Publish("foo", []byte("A"), func() {})
	conn.Publish("foo", []byte("B"), func() {})
	conn.Publish("bar", []byte("C"), func() {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("foo", 0, []byte("A")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("A"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("foo", 1, []byte("B")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("B"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("bar", 2, []byte("C")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("C"))
}

func TestConnection_PublishRetryOnReconnect(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	conn.Publish("foo", []byte("A"), func() {})
	conn.Publish("foo", []byte("B"), func() {})
	conn.Publish("bar", []byte("C"), func() {})

	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("foo", 0, []byte("A")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("A"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("foo", 1, []byte("B")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("B"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("bar", 2, []byte("C")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("C"))

	// Reconnect before ACK'ing. Expect to receive the messages again.
	conn.Reconnect()
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("foo", 0, []byte("A")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("A"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("foo", 1, []byte("B")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("B"))
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("bar", 2, []byte("C")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("C"))

	// ACK the first 2 messages only.
	fakeConn.Push(utils.EncodeACKMessage(1))
	assert.Nil(t, conn.Recv())

	// Reconnect again and now should only get the only unACK'ed message resent.
	conn.Reconnect()
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodePublishMessagePrefix("bar", 2, []byte("C")))
	assert.Equal(t, fakeConn.NextWritten(), []byte("C"))

	// ACK the final message. Now when reconnecting no publishes should be
	// retried.
	fakeConn.Push(utils.EncodeACKMessage(2))
	assert.Nil(t, conn.Recv())
	conn.Reconnect()
	// assert.True(t, fakeConn.NextWritten() == nil) TODO(AD)
}

func TestConnection_OnMessage(t *testing.T) {
	conn, fakeConn := newFakeConnection()
	defer conn.Close()

	messages := []*Message{}

	// Add attachment.
	conn.Attach("foo", func() {}, func(m *Message) {
		data := make([]byte, 0, len(m.Data))
		for _, b := range m.Data {
			data = append(data, b)
		}

		messages = append(messages, &Message{
			Data:   data,
			Offset: m.Offset,
		})
	})
	assert.Equal(t, fakeConn.NextWritten(), utils.EncodeAttachMessage("foo"))
	fakeConn.Push(utils.EncodeAttachedMessage("foo", 0xff))
	assert.Nil(t, conn.Recv())

	fakeConn.Push(utils.EncodeDataMessage("foo", 0x105, []byte("A")))
	assert.Nil(t, conn.Recv())
	fakeConn.Push(utils.EncodeDataMessage("foo", 0x110, []byte("B")))
	assert.Nil(t, conn.Recv())
	// Another topic message should be ignored.
	fakeConn.Push(utils.EncodeDataMessage("bar", 0x102, []byte("C")))
	assert.Nil(t, conn.Recv())
	fakeConn.Push(utils.EncodeDataMessage("foo", 0x115, []byte("D")))
	assert.Nil(t, conn.Recv())

	assert.Equal(t, []*Message{
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

type fakeDialer struct {
	conn net.Conn
}

func (d *fakeDialer) Dial(network string, address string) (net.Conn, error) {
	return d.conn, nil
}

func newFakeConnection() (*connection, *utils.FakeConn) {
	fakeConn := utils.NewFakeConn()
	dialer := &fakeDialer{
		conn: fakeConn,
	}
	opts := defaultOptions("1.2.3.4:123")
	opts.Dialer = dialer

	conn := newConnection(nil, opts)
	conn.Connect()
	return conn, fakeConn
}
