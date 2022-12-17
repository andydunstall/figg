package figg

import (
	"errors"
	"io"
	"net"
	"sync"

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
	opts *Options

	// mu is a mutex protecting the below fields (only locked if the fields
	// are swapped out, should not be held during IO).
	mu sync.Mutex

	conn net.Conn
	// reader reads bytes from the connection.
	reader *reader
}

func newConnection(opts *Options) *connection {
	return &connection{
		opts:   opts,
		mu:     sync.Mutex{},
		conn:   nil,
		reader: nil,
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

	c.connect(conn)

	return nil
}

// Read bytes from the connection. Must only be called from a single goroutine.
func (c *connection) Read() ([]byte, error) {
	if c.reader == nil {
		return nil, ErrNotConnected
	}

	b, err := c.reader.Read()
	if err != nil {
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
	return c.disconnect()
}

func (c *connection) connect(conn net.Conn) {
	c.opts.Logger.Debug("connected", zap.String("addr", c.opts.Addr))

	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.reader = newReader(conn, c.opts.ReadBufLen)
}

func (c *connection) disconnect() error {
	c.opts.Logger.Debug("disconnected", zap.String("addr", c.opts.Addr))

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = nil
	c.reader = nil

	return err
}
