package figg

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andydunstall/figg/utils"
	"go.uber.org/zap"
)

var (
	ErrNotConnected = errors.New("not connected")
)

type reader struct {
	r   io.Reader
	buf []byte
}

func newReader(r io.Reader, bufLen int) *reader {
	return &reader{
		r:   r,
		buf: make([]byte, bufLen),
	}
}

// Reads bytes from the reader. Must only call from one goroutine.
func (r *reader) Read() ([]byte, error) {
	n, err := r.r.Read(r.buf)
	if err != nil {
		return nil, err
	}
	return r.buf[:n], err
}

type connection struct {
	onStateChange func(state ConnState)
	opts          *Options

	attachments     *attachments
	pendingMessages *pendingMessages

	// shutdown is an atomic flag indicating if the client has been shutdown.
	shutdown int32
	// done is a channel used to interrupt reconnect backoff.
	done chan interface{}

	// mu is a mutex protecting the below fields (only locked if the fields
	// are swapped out, should not be held during IO).
	mu sync.Mutex

	conn net.Conn
	// reader reads bytes from the connection.
	reader *reader
	// pending contains bytes read from the connection that have not been
	// processed.
	pending []byte
}

func newConnection(onStateChange func(state ConnState), opts *Options) *connection {
	return &connection{
		onStateChange:   onStateChange,
		opts:            opts,
		attachments:     newAttachments(),
		pendingMessages: newPendingMessages(),
		shutdown:        0,
		done:            make(chan interface{}),
		mu:              sync.Mutex{},
		conn:            nil,
		reader:          nil,
		pending:         []byte{},
	}
}

func (c *connection) Connect() error {
	conn, err := c.opts.Dialer.Dial("tcp", c.opts.Addr)
	if err != nil {
		c.opts.Logger.Error(
			"connection failed",
			zap.String("addr", c.opts.Addr),
			zap.Error(err),
		)
		return err
	}

	c.onConnect(conn)

	return nil
}

func (c *connection) Publish(name string, data []byte, onACK func()) {
	seqNum := c.pendingMessages.Add(name, data, onACK)

	// Ignore any errors as we'll resend on reconnect.
	// Look at using net.Buffers when data large to avoid copying into
	// message buffer.
	c.conn.Write(utils.EncodePublishMessage(name, seqNum, data))
}

func (c *connection) Attach(name string, onAttached func(), onMessage MessageCB) error {
	// Register for an ATTACHED response. Note if sending the ATTACH message
	// fails (eg due to disconnecting), we'll retry all registed attaching
	// attachments.
	if err := c.attachments.AddAttaching(name, onAttached, onMessage); err != nil {
		return err
	}

	// Ignore any errors as we'll resend on reconnect.
	c.conn.Write(utils.EncodeAttachMessage(name))
	return nil
}

func (c *connection) AttachFromOffset(name string, offset uint64, onAttached func(), onMessage MessageCB) error {
	// Register for an ATTACHED response. Note if sending the ATTACH message
	// fails (eg due to disconnecting), we'll retry all registed attaching
	// attachments.
	if err := c.attachments.AddAttachingFromOffset(name, offset, onAttached, onMessage); err != nil {
		return err
	}

	// Ignore any errors as we'll resend on reconnect.
	c.conn.Write(utils.EncodeAttachFromOffsetMessage(name, offset))
	return nil
}

func (c *connection) Detach(name string) {
	// Only send DETACH if we're attached or attaching.
	if c.attachments.AddDetaching(name) {
		// Ignore any errors as we'll resend on reconnect.
		c.conn.Write(utils.EncodeDetachMessage(name))
	}
}

// Read bytes from the connection. Must only be called from a single goroutine.
func (c *connection) Recv() error {
	// TODO(AD) not protected
	if c.reader == nil {
		return ErrNotConnected
	}

	pendingRemaining := c.processPending()

	b, err := c.reader.Read()
	if err != nil {
		// Avoid logging if we are shutdown.
		if s := atomic.LoadInt32(&c.shutdown); s == 1 {
			return err
		}
		c.opts.Logger.Warn(
			"connection closed unexpectedly",
			zap.String("addr", c.opts.Addr),
			zap.Error(err),
		)
		c.onDisconnect()
		return err
	}

	// If there are pending bytes to process, append and process in the next
	// loop.
	if pendingRemaining {
		c.mu.Lock()
		c.pending = append(c.pending, b...)
		c.mu.Unlock()
		return nil
	}

	messageType, payloadLen, ok := utils.DecodeHeader(b)
	if !ok {
		c.mu.Lock()
		c.pending = append(c.pending, b...)
		c.mu.Unlock()
		return nil
	}

	if len(b) < utils.HeaderLen+payloadLen {
		c.mu.Lock()
		c.pending = append(c.pending, b...)
		c.mu.Unlock()
		return nil
	}

	offset := c.onMessage(messageType, utils.HeaderLen, b)
	if offset != len(b) {
		c.mu.Lock()
		c.pending = append(c.pending, b[offset:]...)
		c.mu.Unlock()
	}

	return nil
}

func (c *connection) Reconnect() {
	attempts := 0
	for {
		// If we are shut down give up.
		if s := atomic.LoadInt32(&c.shutdown); s == 1 {
			return
		}

		conn, err := c.opts.Dialer.Dial("tcp", c.opts.Addr)
		if err == nil {
			c.opts.Logger.Debug("reconnect ok", zap.String("addr", c.opts.Addr))
			c.onConnect(conn)
			return
		}

		attempts += 1
		backoff := c.opts.ReconnectBackoffCB(attempts)

		c.opts.Logger.Error(
			"reconnect failed",
			zap.String("addr", c.opts.Addr),
			zap.Int("attempts", attempts),
			zap.Int64("backoff", backoff.Milliseconds()),
			zap.Error(err),
		)

		// If the connection is closed exit immediately.
		select {
		case <-time.After(backoff):
			continue
		case <-c.done:
			return
		}
	}
}

func (c *connection) Close() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&c.shutdown, 1)

	close(c.done)

	return c.onDisconnect()
}

func (c *connection) onMessage(messageType utils.MessageType, offset int, b []byte) int {
	switch messageType {
	case utils.TypeAttached:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)
		topicOffset, offset := utils.DecodeUint64(b, offset)

		c.attachments.OnAttached(topicName, topicOffset)
		return offset
	case utils.TypeDetached:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)

		c.attachments.OnDetached(topicName)
		return offset
	case utils.TypeACK:
		seqNum, offset := utils.DecodeUint64(b, offset)

		c.pendingMessages.Acknowledge(seqNum)
		return offset
	case utils.TypeData:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)
		topicOffset, offset := utils.DecodeUint64(b, offset)
		dataLen, offset := utils.DecodeUint32(b, offset)
		data := b[offset : offset+int(dataLen)]
		offset += int(dataLen)

		c.attachments.OnMessage(topicName, Message{
			Offset: topicOffset,
			Data:   data,
		})
		return offset
	}

	return 0
}

func (c *connection) onConnect(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.reader = newReader(conn, c.opts.ReadBufLen)

	if c.onStateChange != nil {
		c.onStateChange(CONNECTED)
	}

	for _, att := range c.attachments.Attaching() {
		if att.FromOffset {
			c.conn.Write(utils.EncodeAttachFromOffsetMessage(att.Name, att.Offset))
		} else {
			c.conn.Write(utils.EncodeAttachMessage(att.Name))
		}
	}

	for _, att := range c.attachments.Attached() {
		c.conn.Write(utils.EncodeAttachFromOffsetMessage(att.Name, att.Offset))
	}

	for _, topic := range c.attachments.Detaching() {
		c.conn.Write(utils.EncodeDetachMessage(topic))
	}

	for _, m := range c.pendingMessages.Messages() {
		// Look at using net.Buffers when data large to avoid copying into
		// message buffer.
		c.conn.Write(utils.EncodePublishMessage(m.Topic, m.SeqNum, m.Data))
	}
}

func (c *connection) onDisconnect() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = nil
	c.reader = nil

	if c.onStateChange != nil {
		c.onStateChange(DISCONNECTED)
	}

	return err
}

func (c *connection) processPending() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.pending) == 0 {
		return false
	}

	messageType, payloadLen, ok := utils.DecodeHeader(c.pending)
	if !ok {
		return true
	}

	if len(c.pending) < utils.HeaderLen+payloadLen {
		return true
	}

	offset := c.onMessage(messageType, utils.HeaderLen, c.pending)
	c.pending = c.pending[offset:]

	return len(c.pending) == 0
}
