package utils

import (
	"net"
)

type TCPConnection struct {
	conn   net.Conn
	stream *StreamBuffer
}

func NewTCPConnection(conn net.Conn) Connection {
	return &TCPConnection{
		conn:   conn,
		stream: NewStreamBuffer(),
	}
}

func TCPConnect(addr string) (Connection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPConnection{
		conn:   conn,
		stream: NewStreamBuffer(),
	}, nil
}

func (c *TCPConnection) Send(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *TCPConnection) Recv() (MessageType, []byte, error) {
	buf := make([]byte, 1024)
	for {
		messageType, payload, ok := c.stream.Next()
		if ok {
			return messageType, payload, nil
		}

		n, err := c.conn.Read(buf)
		if err != nil {
			return MessageType(0), nil, err
		}

		c.stream.Push(buf[:n])
	}
}

func (c *TCPConnection) Close() error {
	return c.conn.Close()
}
