package conn

import (
	"encoding/binary"
)

type ProtocolConnection struct {
	transport Transport
}

func NewProtocolConnection(transport Transport) Connection {
	conn := &ProtocolConnection{
		transport: transport,
	}

	return conn
}

func (c *ProtocolConnection) Send(offset uint64, m []byte) error {
	b := make([]byte, 8+len(m))
	binary.BigEndian.PutUint64(b, offset)
	for i := 0; i != len(m); i++ {
		b[i+8] = m[i]
	}
	return c.transport.Send(b)
}

func (c *ProtocolConnection) Recv() ([]byte, error) {
	return c.transport.Recv()
}

func (c *ProtocolConnection) Close() error {
	return c.transport.Close()
}
