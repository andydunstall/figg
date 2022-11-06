package sdk

import (
	"testing"

	"github.com/andydunstall/figg/sdk"
	"github.com/andydunstall/figg/wcm/sdk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTopic_PublishSubscribe(t *testing.T) {
	cluster, err := wcm.NewCluster()
	assert.Nil(t, err)
	defer cluster.Shutdown()

	node, err := cluster.AddNode()
	assert.Nil(t, err)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	client, err := figg.NewFigg(&figg.Config{
		Addr:   node.Addr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	messageSubscriber := figg.NewChannelMessageSubscriber()
	client.Subscribe("foo", messageSubscriber)

	client.Publish("foo", []byte("bar"))

	<-messageSubscriber.Ch()
}
