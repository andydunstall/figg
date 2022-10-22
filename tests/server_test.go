package tests

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestServer_PublishAndSubscribe(t *testing.T) {
	s := NewServer()
	defer s.Shutdown()

	addr, err := s.Run()
	assert.Nil(t, err)

	// Add a websocket subscription.
	ws, err := createWS(addr, "foo")
	assert.Nil(t, err)
	defer ws.Close()

	// Publish via REST.
	err = postMessage(addr, "foo", "bar")
	assert.Nil(t, err)

	// Verify we received the message on the websocket subscription.
	mt, message, err := ws.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, websocket.BinaryMessage, mt)
	assert.Equal(t, "bar", string(message))
}
