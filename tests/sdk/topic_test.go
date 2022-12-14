package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/andydunstall/figg/fcm/sdk"
	"github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTopic_PublishSubscribe(t *testing.T) {
	fcmClient := fcm.NewFCM()

	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	client, err := figg.NewFigg(&figg.Config{
		Addr:   cluster.Nodes[0].ProxyAddr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	messageCh := make(chan []byte, 1)
	client.Subscribe(context.Background(), "foo", func(topic string, m []byte) {
		messageCh <- m
	})

	client.Publish(context.Background(), "foo", []byte("bar"))

	assert.Equal(t, []byte("bar"), <-messageCh)
}

// Tests a subscriber recovers lost messages after a disconnect.
func TestTopic_ResumeAfterDisconnect(t *testing.T) {
	fcmClient := fcm.NewFCM()

	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Connect the publisher without the proxy as we can disconnect the
	// subscriber but not the publisher.
	publisherClient, err := figg.NewFigg(&figg.Config{
		Addr:   cluster.Nodes[0].Addr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer publisherClient.Shutdown()

	subscriberClient, err := figg.NewFigg(&figg.Config{
		Addr:   cluster.Nodes[0].ProxyAddr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer subscriberClient.Shutdown()

	messageCh := make(chan []byte, 64)
	subscriberClient.Subscribe(context.Background(), "foo", func(topic string, m []byte) {
		messageCh <- m
	})

	// Publish a message and wait to receive.
	publisherClient.Publish(context.Background(), "foo", []byte("bar"))
	assert.Equal(t, []byte("bar"), <-messageCh)

	// Disable the networking for the subscriber and reenable after 5 seconds.
	// While disconnected publish 3 messages.
	fcmClient.AddChaosPartition(cluster.Nodes[0].ID, fcm.ChaosConfig{
		Duration: 5,
	})

	for i := 0; i < 5; i++ {
		publisherClient.Publish(context.Background(), "foo", []byte(fmt.Sprintf("msg-%d", i)))
	}

	// Once we reconnect we should get the messages we missed.
	for i := 0; i < 5; i++ {
		assert.Equal(t, []byte(fmt.Sprintf("msg-%d", i)), <-messageCh)
	}
}
