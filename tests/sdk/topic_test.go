package sdk

import (
	"fmt"
	"testing"
	"time"

	"github.com/andydunstall/figg/fcm/sdk"
	"github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTopic_PublishSubscribe(t *testing.T) {
	cluster, err := fcm.NewCluster()
	assert.Nil(t, err)
	defer cluster.Shutdown()

	node, err := cluster.AddNode()
	assert.Nil(t, err)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	client, err := figg.NewFigg(&figg.Config{
		Addr:   node.ProxyAddr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	messageSubscriber := figg.NewChannelMessageSubscriber()
	client.Subscribe("foo", messageSubscriber)

	client.Publish("foo", []byte("bar"))

	assert.Equal(t, []byte("bar"), <-messageSubscriber.Ch())
}

// Tests a subscriber recovers lost messages after a disconnect.
func TestTopic_ResumeAfterDisconnect(t *testing.T) {
	cluster, err := fcm.NewCluster()
	assert.Nil(t, err)
	defer cluster.Shutdown()

	node, err := cluster.AddNode()
	assert.Nil(t, err)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Connect the publisher without the proxy as we can disconnect the
	// subscriber but not the publisher.
	publisherClient, err := figg.NewFigg(&figg.Config{
		Addr:   node.Addr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer publisherClient.Shutdown()

	subscriberClient, err := figg.NewFigg(&figg.Config{
		Addr:   node.ProxyAddr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer subscriberClient.Shutdown()

	messageSubscriber := figg.NewChannelMessageSubscriber()
	subscriberClient.Subscribe("foo", messageSubscriber)

	// TODO(AD) Wait for subscribe to become attached.
	<-time.After(time.Second)

	// Publish a message and wait to receive.
	publisherClient.Publish("foo", []byte("bar"))
	assert.Equal(t, []byte("bar"), <-messageSubscriber.Ch())

	// Disable the networking for the subscriber and reenable after 5 seconds.
	// While disconnected publish 3 messages.
	assert.Nil(t, node.Disable())
	go func() {
		<-time.After(3 * time.Second)
		assert.Nil(t, node.Enable())
	}()

	for i := 0; i < 5; i++ {
		publisherClient.Publish("foo", []byte(fmt.Sprintf("msg-%d", i)))
	}

	// Once we reconnect we should get the messages we missed.
	for i := 0; i < 5; i++ {
		assert.Equal(t, []byte(fmt.Sprintf("msg-%d", i)), <-messageSubscriber.Ch())
	}
}
