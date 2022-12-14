package utils

type Connection interface {
	Send(b []byte) error
	Recv() (MessageType, []byte, error)
	Close() error
}
