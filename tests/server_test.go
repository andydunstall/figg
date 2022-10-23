package tests

import (
	"fmt"
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
	ws, err := createWS(addr, "foo", "")
	assert.Nil(t, err)
	defer ws.Close()

	// Publish via REST.
	for i := 0; i != 10; i++ {
		assert.Nil(t, postMessage(addr, "foo", fmt.Sprintf("%d", i)))
	}

	// Verify we received the messages on the websocket subscription.
	for i := 0; i != 10; i++ {
		mt, message, err := ws.ReadMessage()
		assert.Nil(t, err)
		assert.Equal(t, websocket.BinaryMessage, mt)
		assert.Equal(t, fmt.Sprintf("%d", i), string(message))
	}
}

func TestServer_PublishAndSubscribeFromOffset(t *testing.T) {
	s := NewServer()
	defer s.Shutdown()

	addr, err := s.Run()
	assert.Nil(t, err)

	// Publish 5 messages via REST, prior to subscribing.
	for i := 0; i != 5; i++ {
		assert.Nil(t, postMessage(addr, "foo", fmt.Sprintf("%d", i)))
	}

	// Add a websocket subscription.
	ws, err := createWS(addr, "foo", "offset=0")
	assert.Nil(t, err)
	defer ws.Close()

	// Publish another 5 messages via REST, after subscribing.
	for i := 5; i != 10; i++ {
		assert.Nil(t, postMessage(addr, "foo", fmt.Sprintf("%d", i)))
	}

	// Verify we received the messages on the websocket subscription.
	for i := 0; i != 10; i++ {
		mt, message, err := ws.ReadMessage()
		assert.Nil(t, err)
		assert.Equal(t, websocket.BinaryMessage, mt)
		assert.Equal(t, fmt.Sprintf("%d", i), string(message))
	}
}
