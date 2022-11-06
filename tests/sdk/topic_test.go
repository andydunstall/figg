package sdk

import (
	"testing"
	"time"

	"github.com/andydunstall/wombat/sdk"
	"github.com/andydunstall/wombat/wcm/sdk"
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
	client, err := wombat.NewWombat(&wombat.Config{
		Addr:   node.Addr,
		Logger: logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	// TODO(AD) wait to connect
	<-time.After(time.Second)

	messageSubscriber := wombat.NewChannelMessageSubscriber()
	client.Subscribe("foo", messageSubscriber)

	client.Publish("foo", []byte("bar"))

	<-messageSubscriber.Ch()
}
