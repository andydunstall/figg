package tests

import (
	"fmt"
	"testing"

	"github.com/andydunstall/wombat/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestServer_PublishAndSubscribe(t *testing.T) {
	s := NewServer()
	defer s.Shutdown()

	addr, err := s.Run()
	assert.Nil(t, err)

	// Add a websocket subscription.
	client, err := client.NewWombat(addr, "foo")
	assert.Nil(t, err)
	defer client.Shutdown()

	// Publish via REST.
	for i := 0; i != 10; i++ {
		assert.Nil(t, postMessage(addr, "foo", fmt.Sprintf("%d", i)))
	}

	// Verify we received the messages on the websocket subscription.
	for i := 0; i != 10; i++ {
		message := <-client.MessagesCh()
		assert.Equal(t, fmt.Sprintf("%d", i), string(message))
	}
}
