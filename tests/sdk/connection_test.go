package sdk

import (
	"testing"
	"time"

	"github.com/andydunstall/figg/fcm/sdk"
	"github.com/andydunstall/figg/sdk"
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
	fcmClient := fcm.NewFCM()

	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	stateSubscriber := figg.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            cluster.Nodes[0].ProxyAddr,
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
	fcmClient := fcm.NewFCM()

	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	// Disable the networking for the node and reenable after 5 seconds.
	fcmClient.AddChaosPartition(cluster.Nodes[0].ID, fcm.ChaosConfig{
		Duration: 5,
	})

	stateSubscriber := figg.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            cluster.Nodes[0].ProxyAddr,
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
	fcmClient := fcm.NewFCM()

	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	stateSubscriber := figg.NewChannelStateSubscriber()
	logger, _ := zap.NewDevelopment()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            cluster.Nodes[0].Addr,
		StateSubscriber: stateSubscriber,
		Logger:          logger,
	})
	assert.Nil(t, err)
	defer client.Shutdown()

	evt, ok := waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateConnected, evt)

	// Disable the networking for the node and reenable after 3 second.
	fcmClient.AddChaosPartition(cluster.Nodes[0].ID, fcm.ChaosConfig{
		Duration: 3,
	})

	evt, ok = waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateDisconnected, evt)

	evt, ok = waitForStateWithTimeout(stateSubscriber, 5*time.Second)
	assert.True(t, ok)
	assert.Equal(t, figg.StateConnected, evt)
}
