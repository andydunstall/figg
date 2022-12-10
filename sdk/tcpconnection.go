package figg

import (
	"encoding/binary"
	"net"
)

type TCPConnection struct {
	conn   net.Conn
	stream *StreamBuffer
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
	prefix := make([]byte, 4)
	binary.BigEndian.PutUint32(prefix, uint32(len(b)))

	bufs := net.Buffers{}
	bufs = append(bufs, prefix)
	bufs = append(bufs, b)
	_, err := bufs.WriteTo(c.conn)
	return err
}

func (c *TCPConnection) Recv() ([]byte, error) {
	buf := make([]byte, 1024)
	for {
		b, ok := c.stream.Next()
		if ok {
			return b, nil
		}

		n, err := c.conn.Read(buf)
		if err != nil {
			return nil, err
		}

		c.stream.Push(buf[:n])
	}
}

func (c *TCPConnection) Close() error {
	return c.conn.Close()
}
