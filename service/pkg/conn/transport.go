package conn

type Transport interface {
	Send(m *ProtocolMessage) error
	Recv() (*ProtocolMessage, error)
	Close() error
}
