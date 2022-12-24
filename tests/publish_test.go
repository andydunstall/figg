package tests

import (
	"fmt"
	"testing"
	"time"

	fcm "github.com/andydunstall/figg/fcm/lib"
	figg "github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
)

// Tests the publisher resends messages when it reconnects to the server
// following a disconnect.
func TestPublish_ResendAfterDisconnect(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	// Note the publisher uses the proxy address (which is disconnected), but
	// the subscriber uses the nodes address (which is not disconnected).

	subClient, err := figg.Connect(node.Addr)
	assert.Nil(t, err)
	defer subClient.Close()

	pubClient, err := figg.Connect(node.ProxyAddr, figg.WithLogger(setupLogger()))
	assert.Nil(t, err)
	defer pubClient.Close()

	// Add a buffer so the subscribe callback doesn't block.
	messagesCh := make(chan *figg.Message, 25)
	assert.Nil(t, subClient.Subscribe("foo", func(m *figg.Message) {
		messagesCh <- m
	}))

	go func() {
		for i := 0; i != 25; i++ {
			pubClient.Publish("foo", []byte(fmt.Sprintf("message-%d", i)))
			// Add a delay so most publishes are sent while the subscriber
			// is disconnected.
			<-time.After(time.Millisecond * 250)
		}
	}()

	// Disable the networking for the node and reenable after 2 second. Note
	// this only affects the publisher given only it is connected to the proxy.
	node.PartitionFor(2)

	for i := 0; i != 25; i++ {
		m := <-messagesCh
		assert.Equal(t, fmt.Sprintf("message-%d", i), string(m.Data))
	}
}
