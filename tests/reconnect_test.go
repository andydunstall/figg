package tests

import (
	"testing"

	fcm "github.com/andydunstall/figg/fcm/lib"
	figg "github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestReconnect_ReconnectAfterConnectionDrops(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	// Add buffered chan to avoid blocking.
	connStateChan := make(chan figg.ConnState, 1)

	client, err := figg.Connect(
		node.ProxyAddr,
		figg.WithLogger(setupLogger()),
		figg.WithConnStateChangeCB(func(s figg.ConnState) {
			connStateChan <- s
		}),
	)
	assert.Nil(t, err)
	defer client.Close()

	assert.Equal(t, figg.CONNECTED, <-connStateChan)

	// Disable the networking for the node and reenable after 2 second.
	node.PartitionFor(2)

	// The client should detect its disconnected then reconnect.
	assert.Equal(t, figg.DISCONNECTED, <-connStateChan)
	assert.Equal(t, figg.CONNECTED, <-connStateChan)
}

func TestReconnect_ReconnectAfterPacketsDrop(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	// Add buffered chan to avoid blocking.
	connStateChan := make(chan figg.ConnState, 1)

	client, err := figg.Connect(
		node.ProxyAddr,
		figg.WithLogger(setupLogger()),
		figg.WithConnStateChangeCB(func(s figg.ConnState) {
			connStateChan <- s
		}),
	)
	assert.Nil(t, err)
	defer client.Close()

	assert.Equal(t, figg.CONNECTED, <-connStateChan)

	// Drop all packets without closing the connection. This should cause ping
	// to expire and reconnect.
	node.DropActive()

	// The client should detect its disconnected then reconnect.
	assert.Equal(t, figg.DISCONNECTED, <-connStateChan)
	assert.Equal(t, figg.CONNECTED, <-connStateChan)
}
