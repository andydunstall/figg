package figg

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

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

	// shutdown is an atomic flag indicating if the client has been shutdown.
	shutdown int32

	// mu is a mutex protecting the below fields (only locked if the fields
	// are swapped out, should not be held during IO).
	mu sync.Mutex

	conn net.Conn
	// reader reads bytes from the connection.
	reader *reader
}

func newConnection(onStateChange func(state ConnState), opts *Options) *connection {
	return &connection{
		onStateChange: onStateChange,
		opts:          opts,
		shutdown:      0,
		mu:            sync.Mutex{},
		conn:          nil,
		reader:        nil,
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

// Read bytes from the connection. Must only be called from a single goroutine.
func (c *connection) Read() ([]byte, error) {
	if c.reader == nil {
		return nil, ErrNotConnected
	}

	b, err := c.reader.Read()
	if err != nil {
		// Avoid logging if we are shutdown.
		if s := atomic.LoadInt32(&c.shutdown); s == 1 {
			return nil, err
		}
		c.opts.Logger.Warn(
			"connection closed unexpectedly",
			zap.String("addr", c.opts.Addr),
			zap.Error(err),
		)
		return nil, err
	}
	return b, nil
}

func (c *connection) Close() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&c.shutdown, 1)

	return c.onDisconnect()
}

func (c *connection) onConnect(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.reader = newReader(conn, c.opts.ReadBufLen)

	c.onStateChange(CONNECTED)
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

	c.onStateChange(DISCONNECTED)

	return err
}
