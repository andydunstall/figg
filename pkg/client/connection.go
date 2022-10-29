package client

type Connection interface {
	Recv() ([]byte, error)
	Close() error
}
