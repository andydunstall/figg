package figg

import (
	"errors"
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

type connection struct {
	onStateChange func(state ConnState)
	opts          *Options

	attachments *attachments
	window      *slidingWindow

	// shutdown is an atomic flag indicating if the client has been shutdown.
	shutdown int32
	// done is a channel used to interrupt reconnect backoff.
	done chan interface{}

	// mu is a mutex protecting the below fields. Note should not be held during
	// IO, so if performing IO take a copy then unlock.
	mu sync.Mutex

	conn net.Conn
	// reader reads messages from the connection. Must only be accessed from
	// the read loop.
	reader *utils.BufferedReader
	// writer writes messages to the connection.
	writer *utils.BufferedWriter

	// outstandingPings is the number of pings that have been sent but not
	// acknowledged with a pong.
	outstandingPings int
}

func newConnection(onStateChange func(state ConnState), opts *Options) *connection {
	return &connection{
		onStateChange: onStateChange,
		opts:          opts,
		attachments:   newAttachments(),
		window:        newSlidingWindow(opts.WindowSize),
		shutdown:      0,
		done:          make(chan interface{}),
		mu:            sync.Mutex{},
		conn:          nil,
		reader:        nil,
		writer:        nil,
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
	seqNum := c.window.Push(name, data, onACK)

	c.opts.Logger.Debug(
		"publish",
		zap.String("topic", name),
		zap.Int("data-len", len(data)),
		zap.Uint64("seqNum", seqNum),
	)

	// Ignore any errors as we'll resend on reconnect.
	// Look at using net.Buffers when data large to avoid copying into
	// message buffer.
	c.send(
		utils.EncodePublishMessagePrefix(name, seqNum, data),
		data,
	)
}

func (c *connection) Attach(name string, onAttached func(), onMessage MessageCB) error {
	c.opts.Logger.Debug("attach", zap.String("topic", name))

	// Register for an ATTACHED response. Note if sending the ATTACH message
	// fails (eg due to disconnecting), we'll retry all registed attaching
	// attachments.
	if err := c.attachments.AddAttaching(name, onAttached, onMessage); err != nil {
		return err
	}

	// Ignore any errors as we'll resend on reconnect.
	c.send(utils.EncodeAttachMessage(name))
	return nil
}

func (c *connection) AttachFromOffset(name string, offset uint64, onAttached func(), onMessage MessageCB) error {
	c.opts.Logger.Debug(
		"attach from offset",
		zap.String("topic", name),
		zap.Uint64("offset", offset),
	)

	// Register for an ATTACHED response. Note if sending the ATTACH message
	// fails (eg due to disconnecting), we'll retry all registed attaching
	// attachments.
	if err := c.attachments.AddAttachingFromOffset(name, offset, onAttached, onMessage); err != nil {
		return err
	}

	// Ignore any errors as we'll resend on reconnect.
	c.send(utils.EncodeAttachFromOffsetMessage(name, offset))
	return nil
}

func (c *connection) Detach(name string) {
	c.opts.Logger.Debug(
		"detach",
		zap.String("topic", name),
	)

	// Only send DETACH if we're attached or attaching.
	if c.attachments.AddDetaching(name) {
		// Ignore any errors as we'll resend on reconnect.
		c.send(utils.EncodeDetachMessage(name))
	}
}

func (c *connection) Ping() error {
	timestamp := uint64(time.Now().UnixNano())

	c.opts.Logger.Debug(
		"ping",
		zap.Uint64("timestamp", timestamp),
	)

	c.mu.Lock()
	if c.outstandingPings >= c.opts.MaxPingOut {
		c.opts.Logger.Warn(
			"connection closed: ping expired",
			zap.String("addr", c.opts.Addr),
		)
		c.mu.Unlock()
		c.onDisconnect()
		return ErrNotConnected
	}
	c.mu.Unlock()

	c.send(utils.EncodePingMessage(timestamp))

	c.mu.Lock()
	c.outstandingPings++
	c.mu.Unlock()

	return nil
}

// Read bytes from the connection. Must only be called from a single goroutine.
func (c *connection) Recv() error {
	// TODO(AD) not protected (written on reconnect)
	if c.reader == nil {
		return ErrNotConnected
	}

	messageType, payload, err := c.reader.Read()
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

	c.onMessage(messageType, payload)

	return nil
}

func (c *connection) Reconnect() {
	c.opts.Logger.Debug("reconnect")

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
	if c.writer != nil {
		c.writer.Close()
	}

	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&c.shutdown, 1)

	close(c.done)

	return c.onDisconnect()
}

func (c *connection) send(bufs ...[]byte) error {
	// Copy to avoid locking during IO.
	c.mu.Lock()
	writer := c.writer
	c.mu.Unlock()

	if writer == nil {
		return ErrNotConnected
	}
	return writer.Write(bufs...)
}

func (c *connection) onMessage(messageType utils.MessageType, b []byte) int {
	offset := 0
	switch messageType {
	case utils.TypeAttached:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)
		topicOffset, offset := utils.DecodeUint64(b, offset)

		c.opts.Logger.Debug(
			"on message",
			zap.String("message-type", messageType.String()),
			zap.String("topic", topicName),
			zap.Uint64("offset", topicOffset),
		)

		c.attachments.OnAttached(topicName, topicOffset)
		return offset
	case utils.TypeDetached:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)

		c.opts.Logger.Debug(
			"on message",
			zap.String("message-type", messageType.String()),
			zap.String("topic", topicName),
		)

		c.attachments.OnDetached(topicName)
		return offset
	case utils.TypeACK:
		seqNum, offset := utils.DecodeUint64(b, offset)

		c.opts.Logger.Debug(
			"on message",
			zap.String("message-type", messageType.String()),
			zap.Uint64("seq-num", seqNum),
		)

		c.window.Acknowledge(seqNum)
		return offset
	case utils.TypeData:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)
		topicOffset, offset := utils.DecodeUint64(b, offset)
		dataLen, offset := utils.DecodeUint32(b, offset)
		data := make([]byte, dataLen)
		copy(data, b[offset:offset+int(dataLen)])
		offset += int(dataLen)

		c.opts.Logger.Debug(
			"on message",
			zap.String("message-type", messageType.String()),
			zap.String("topic", topicName),
			zap.Uint64("offset", topicOffset),
			zap.Int("data-len", len(data)),
		)

		c.attachments.OnMessage(topicName, &Message{
			Offset: topicOffset,
			Data:   data,
		})
		return offset
	case utils.TypePong:
		timestamp, _ := utils.DecodeUint64(b, offset)

		c.opts.Logger.Debug(
			"on message",
			zap.String("message-type", messageType.String()),
			zap.Duration("rtt", time.Duration(uint64(time.Now().UnixNano())-timestamp)),
		)

		c.mu.Lock()
		c.outstandingPings--
		c.mu.Unlock()

		return offset
	}

	return 0
}

func (c *connection) onConnect(conn net.Conn) {
	c.setNetConn(conn)

	if c.onStateChange != nil {
		c.onStateChange(CONNECTED)
	}

	for _, att := range c.attachments.Attaching() {
		if att.FromOffset {
			c.send(utils.EncodeAttachFromOffsetMessage(att.Name, att.Offset))
		} else {
			c.send(utils.EncodeAttachMessage(att.Name))
		}
	}

	for _, att := range c.attachments.Attached() {
		c.opts.Logger.Debug(
			"re-attach",
			zap.String("topic", att.Name),
			zap.Uint64("offset", att.Offset),
		)

		c.send(utils.EncodeAttachFromOffsetMessage(att.Name, att.Offset))
	}

	for _, topic := range c.attachments.Detaching() {
		c.opts.Logger.Debug(
			"re-detach",
			zap.String("topic", topic),
		)

		c.send(utils.EncodeDetachMessage(topic))
	}

	for _, m := range c.window.Messages() {
		c.opts.Logger.Debug(
			"re-publish",
			zap.String("topic", m.Topic),
			zap.Int("data-len", len(m.Data)),
			zap.Uint64("seqNum", m.SeqNum),
		)

		// Look at using net.Buffers when data large to avoid copying into
		// message buffer.
		c.send(
			utils.EncodePublishMessagePrefix(m.Topic, m.SeqNum, m.Data),
			m.Data,
		)
	}
}

func (c *connection) onDisconnect() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()

	c.unsetNetConn()

	if c.onStateChange != nil {
		c.onStateChange(DISCONNECTED)
	}

	return err
}

func (c *connection) setNetConn(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.reader = utils.NewBufferedReader(conn, c.opts.ReadBufLen)
	c.writer = utils.NewBufferedWriter(conn)
}

func (c *connection) unsetNetConn() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = nil
	c.reader = nil
	if c.writer != nil {
		c.writer.Close()
	}
	c.writer = nil

	c.outstandingPings = 0
}
