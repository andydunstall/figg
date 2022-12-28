package fcm

import (
	"net"
)

// TODO(AD) Not thread safe.
type Connection struct {
	upstream   net.Conn
	downstream net.Conn
	proxy      *Proxy
	drop bool
}

func NewConnection(upstream net.Conn, downstream net.Conn, proxy *Proxy) *Connection {
	conn := &Connection{
		upstream:   upstream,
		downstream: downstream,
		proxy:      proxy,
		drop: false,
	}
	go conn.forwardDownstream()
	go conn.forwardUpstream()
	return conn
}

func (c *Connection) Drop() {
	c.drop = true
}

func (c *Connection) Close() error {
	if err := c.upstream.Close(); err != nil {
		return err
	}
	if err := c.downstream.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Connection) forwardDownstream() {
	buf := make([]byte, 1<<15)
	for {
		n, err := c.downstream.Read(buf)
		if err != nil {
			return
		}

		if !c.drop {
			c.upstream.Write(buf[:n])
		}
	}
}

func (c *Connection) forwardUpstream() {
	buf := make([]byte, 1<<15)
	for {
		n, err := c.upstream.Read(buf)
		if err != nil {
			return
		}

		if !c.drop {
			c.downstream.Write(buf[:n])
		}
	}
}
