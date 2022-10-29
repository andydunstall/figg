package conn

type ProtocolConnection struct {
	transport Transport
}

func NewProtocolConnection(transport Transport) Connection {
	return &ProtocolConnection{
		transport: transport,
	}
}

func (c *ProtocolConnection) Send(m *ProtocolMessage) error {
	return c.transport.Send(m)
}

func (c *ProtocolConnection) Recv() (*ProtocolMessage, error) {
	return c.transport.Recv()
}

func (c *ProtocolConnection) Close() error {
	return c.transport.Close()
}
