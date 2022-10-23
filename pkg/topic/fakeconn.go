package topic

import (
	"fmt"
)

type FakeConn struct {
	Sent chan Message
}

func NewFakeConn() *FakeConn {
	return &FakeConn{
		Sent: make(chan Message, 64),
	}
}

func (c *FakeConn) Send(offset uint64, m []byte) error {
	c.Sent <- Message{
		Offset:  offset,
		Message: m,
	}
	return nil
}

func (c *FakeConn) Recv() ([]byte, error) {
	return nil, fmt.Errorf("EOF")
}
