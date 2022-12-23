package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests reading all readable bytes (where the read buffer exceeds the
// readable bytes)
func TestFakeConn_ReadAllBytes(t *testing.T) {
	conn := NewFakeConn()
	defer conn.Close()

	conn.Push([]byte("foo"))

	buf := make([]byte, 5)
	n, _ := conn.Read(buf)
	assert.Equal(t, []byte("foo"), buf[:n])
}

func TestFakeConn_ReadPartialBytes(t *testing.T) {
	conn := NewFakeConn()
	defer conn.Close()

	conn.Push([]byte("foo"))

	// Use a small buffer so only one byte can be read at a time.
	buf := make([]byte, 1)

	n, _ := conn.Read(buf)
	assert.Equal(t, []byte("f"), buf[:n])
	n, _ = conn.Read(buf)
	assert.Equal(t, []byte("o"), buf[:n])
	n, _ = conn.Read(buf)
	assert.Equal(t, []byte("o"), buf[:n])
}

func TestFakeConn_NextWritten(t *testing.T) {
	conn := NewFakeConn()
	defer conn.Close()

	conn.Write([]byte("foo"))
	conn.Write([]byte("bar"))

	assert.Equal(t, []byte("foo"), conn.NextWritten())
	assert.Equal(t, []byte("bar"), conn.NextWritten())
}
