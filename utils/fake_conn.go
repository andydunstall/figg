package utils

import (
	"net"
	"sync"
	"time"
)

// FakeConn is a fake network connection used for testing. This is thread safe
// as may be accessed by a backgroud read/write goroutine.
type FakeConn struct {
	cv       *sync.Cond
	readable []byte
	written  [][]byte
}

func NewFakeConn() *FakeConn {
	return &FakeConn{
		readable: []byte{},
		written:  [][]byte{},
		cv:       sync.NewCond(&sync.Mutex{}),
	}
}

func (c *FakeConn) Read(b []byte) (int, error) {
	c.cv.L.Lock()
	defer c.cv.L.Unlock()
	for len(c.readable) == 0 {
		c.cv.Wait()
	}

	// Fill the buffer as much as possible.
	n := len(c.readable)
	if n > len(b) {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		b[i] = c.readable[i]
	}
	c.readable = c.readable[n:]
	return n, nil
}

func (c *FakeConn) Write(b []byte) (n int, err error) {
	c.cv.L.Lock()
	c.written = append(c.written, b)
	c.cv.L.Unlock()

	c.cv.Signal()

	return len(b), nil
}

// NextWritten returns the next write to the connection. This blocks until
// received.
func (c *FakeConn) NextWritten() []byte {
	c.cv.L.Lock()
	for len(c.written) == 0 {
		c.cv.Wait()
	}

	next := c.written[0]
	c.written = c.written[1:]
	c.cv.L.Unlock()
	return next
}

// Push adds a new buffer for read to return.
func (c *FakeConn) Push(b []byte) {
	c.cv.L.Lock()
	c.readable = append(c.readable, b...)
	c.cv.L.Unlock()

	c.cv.Signal()
}

func (c *FakeConn) Close() error {
	return nil
}

// Adding nop methods to satisfy the net.Conn interface.

func (c *FakeConn) LocalAddr() net.Addr {
	return nil
}

func (c *FakeConn) RemoteAddr() net.Addr {
	return nil
}

func (c *FakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *FakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *FakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
