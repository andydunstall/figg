package conn

type Connection interface {
	Send(offset uint64, m []byte) error
	Recv() ([]byte, error)
	Close() error
}
