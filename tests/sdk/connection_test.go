package sdk

import (
	"testing"
	"time"

	"github.com/andydunstall/figg/sdk"
	"github.com/andydunstall/figg/fcm/sdk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func waitForStateWithTimeout(stateSubscriber *figg.ChannelStateSubscriber, timeout time.Duration) (figg.State, bool) {
	select {
	case <-time.After(timeout):
		return figg.State(0), false
	case state := <-stateSubscriber.Ch():
		return state, true
	}
}

// Tests the SDK connects when figg is reachable.
func TestConnection_Connect(t *testing.T) {
	cluster, err := fcm.NewCluster()
	assert.Nil(t, err)
	defer cluster.Shutdown()

	node, err := cluster.AddNode()
	assert.Nil(t, err)

	stateSubscriber := figg.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            node.Addr,
		StateSubscriber: stateSubscriber,
		Logger:          logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	evt, ok := waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateConnected, evt)
}

// Tests if figg is unreachable when the SDK initally tries to connect, it
// retries and succeeds once figg is reachable.
func TestConnection_ConnectOnceReachable(t *testing.T) {
	cluster, err := fcm.NewCluster()
	assert.Nil(t, err)
	defer cluster.Shutdown()

	node, err := cluster.AddNode()
	assert.Nil(t, err)

	// Disable the networking for the node and reenable after 5 seconds.
	assert.Nil(t, node.Disable())
	go func() {
		<-time.After(5 * time.Second)
		assert.Nil(t, node.Enable())
	}()

	stateSubscriber := figg.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            node.Addr,
		StateSubscriber: stateSubscriber,
		Logger:          logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	evt, ok := waitForStateWithTimeout(stateSubscriber, 10*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateConnected, evt)
}

// Tests if the connection to figg is disconnected the SDK detects the
// disconnection and tries to reconnect.
func TestConnection_ReconnectAfterDisconnected(t *testing.T) {
	cluster, err := fcm.NewCluster()
	assert.Nil(t, err)
	defer cluster.Shutdown()

	node, err := cluster.AddNode()
	assert.Nil(t, err)

	stateSubscriber := figg.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            node.Addr,
		StateSubscriber: stateSubscriber,
		Logger:          logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	evt, ok := waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateConnected, evt)

	// Disable the networking for the node and reenable after 1 second.
	assert.Nil(t, node.Disable())
	go func() {
		<-time.After(1 * time.Second)
		assert.Nil(t, node.Enable())
	}()

	evt, ok = waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateDisconnected, evt)

	evt, ok = waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateConnected, evt)
}
