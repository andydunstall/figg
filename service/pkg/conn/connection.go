package conn

type Connection interface {
	Send(b []byte) error
	Recv() ([]byte, error)
	Close() error
}
