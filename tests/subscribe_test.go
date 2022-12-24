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
