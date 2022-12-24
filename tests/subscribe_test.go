package tests

import (
	"fmt"
	"testing"
	"time"

	fcm "github.com/andydunstall/figg/fcm/lib"
	figg "github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestSubscribe_SubscribeToTopics(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	subClient, err := figg.Connect(node.Addr, figg.WithLogger(setupLogger()))
	assert.Nil(t, err)
	defer subClient.Close()

	pubClient, err := figg.Connect(node.Addr, figg.WithLogger(setupLogger()))
	assert.Nil(t, err)
	defer pubClient.Close()

	// Add a buffer so the subscribe callback doesn't block.
	messagesCh := make(chan *figg.Message, 10)
	assert.Nil(t, subClient.Subscribe("foo", func(m *figg.Message) {
		messagesCh <- m
	}))

	for i := 0; i != 10; i++ {
		pubClient.Publish("foo", []byte(fmt.Sprintf("message-%d", i)))
	}

	for i := 0; i != 10; i++ {
		m := <-messagesCh
		assert.Equal(t, fmt.Sprintf("message-%d", i), string(m.Data))
	}
}

func TestSubscribe_SubscribeFromOffset(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	pubClient, err := figg.Connect(node.Addr, figg.WithLogger(setupLogger()))
	assert.Nil(t, err)
	defer pubClient.Close()

	// Publish all messages before creating a subscriber.
	for i := 0; i != 10; i++ {
		pubClient.Publish("foo", []byte(fmt.Sprintf("message-%d", i)))
	}

	subClient, err := figg.Connect(node.Addr, figg.WithLogger(setupLogger()))
	assert.Nil(t, err)
	defer subClient.Close()

	// Add a buffer so the subscribe callback doesn't block.
	messagesCh := make(chan *figg.Message, 10)
	// Subscribe from offset 0 to get all messages published on the topic.
	assert.Nil(t, subClient.Subscribe("foo", func(m *figg.Message) {
		messagesCh <- m
	}, figg.WithOffset(0)))

	for i := 0; i != 10; i++ {
		m := <-messagesCh
		assert.Equal(t, fmt.Sprintf("message-%d", i), string(m.Data))
	}
}

// Tests the subscriber does not drop messages even if it is disconnected from
// the node (while the publisher is not disconnected).
func TestSubscribe_ResumeAfterDisconnect(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	// Note the subscriber uses the proxy address (which is disconnected), but
	// the publisher uses the nodes address (which is not disconnected).

	subClient, err := figg.Connect(node.ProxyAddr, figg.WithLogger(setupLogger()))
	assert.Nil(t, err)
	defer subClient.Close()

	pubClient, err := figg.Connect(node.Addr)
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

	// Receive 5 messages then disconnect the node.
	for i := 0; i != 5; i++ {
		m := <-messagesCh
		assert.Equal(t, fmt.Sprintf("message-%d", i), string(m.Data))
	}

	// Disable the networking for the node and reenable after 2 second. Note
	// this only affects the subscriber given only it is connected to the proxy.
	node.PartitionFor(2)

	// Expect to receive the rest of the messages after the SDK auto-reconnects.
	for i := 5; i != 25; i++ {
		m := <-messagesCh
		assert.Equal(t, fmt.Sprintf("message-%d", i), string(m.Data))
	}
}

// TODO(AD) Test TestSubscribe_ResumeAfterDisconnect without having received
// any messages (requiring ATTACHED to include an offset). Test a non-zero
// offset.
