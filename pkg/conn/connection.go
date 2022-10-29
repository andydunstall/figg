package conn

type Connection interface {
	Send(m *ProtocolMessage) error
	Recv() (*ProtocolMessage, error)
	Close() error
}
