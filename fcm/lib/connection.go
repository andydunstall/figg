package fcm

import (
	"io"
	"net"
)

type Connection struct {
	upstream   net.Conn
	downstream net.Conn
	proxy      *Proxy
}

func NewConnection(upstream net.Conn, downstream net.Conn, proxy *Proxy) *Connection {
	conn := &Connection{
		upstream:   upstream,
		downstream: downstream,
		proxy:      proxy,
	}
	go conn.forwardDownstream()
	go conn.forwardUpstream()
	return conn
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
	io.Copy(c.downstream, c.upstream)
}

func (c *Connection) forwardUpstream() {
	io.Copy(c.upstream, c.downstream)
}
