package conn

import (
	"fmt"
)

type FakeConnection struct {
	Sent chan Message
}

func NewFakeConnection() *FakeConnection {
	return &FakeConnection{
		Sent: make(chan Message),
	}
}

func (c *FakeConnection) Send(offset uint64, m []byte) error {
	c.Sent <- Message{
		Offset:  offset,
		Message: m,
	}
	return nil
}

func (c *FakeConnection) Recv() ([]byte, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (c *FakeConnection) Close() error {
	return nil
}
