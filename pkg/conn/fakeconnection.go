package conn

import (
	"fmt"
)

type FakeConnection struct {
	Sent chan *ProtocolMessage
}

func NewFakeConnection() *FakeConnection {
	return &FakeConnection{
		Sent: make(chan *ProtocolMessage),
	}
}

func (c *FakeConnection) Send(m *ProtocolMessage) error {
	c.Sent <- m
	return nil
}

func (c *FakeConnection) Recv() (*ProtocolMessage, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (c *FakeConnection) Close() error {
	return nil
}
