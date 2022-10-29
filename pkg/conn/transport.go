package conn

type Transport interface {
	Send(b []byte) error
	Recv() ([]byte, error)
	Close() error
}
