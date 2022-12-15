package figg

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ErrorDialer struct {}

func (d *ErrorDialer) Dial(network string, address string) (net.Conn, error) {
	return nil, errors.New("failed to connect")
}

func TestConnect_ServerUnreachable(t *testing.T) {
	_, err := Connect("127.0.0.1:8119", WithDialer(&ErrorDialer{}))
	assert.Error(t, err)
}
