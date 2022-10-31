package sdk

import (
	"context"
	"testing"
	"time"

	"github.com/andydunstall/wombat/sdk"
	"github.com/andydunstall/wombat/wcm/sdk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func waitForStateWithTimeout(stateSubscriber *wombat.ChannelStateSubscriber, timeout time.Duration) (wombat.State, bool) {
	select {
	case <-time.After(timeout):
		return wombat.State(0), false
	case state := <-stateSubscriber.Ch():
		return state, true
	}
}

func TestConnection_Connect(t *testing.T) {
	wcm, err := wcm.Connect()
	assert.Nil(t, err)
	defer wcm.Close()

	cluster, err := wcm.CreateCluster(context.Background())
	assert.Nil(t, err)
	defer cluster.Close(context.Background())

	node, err := cluster.AddNode(context.Background())
	assert.Nil(t, err)

	stateSubscriber := wombat.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client := wombat.NewWombat(&wombat.Config{
		Addr:            node.Addr,
		StateSubscriber: stateSubscriber,
		Logger:          logger,
	})
	defer client.Shutdown()

	evt, ok := waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, wombat.StateConnected, evt)
}
