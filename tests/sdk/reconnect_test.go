package sdk

import (
	"testing"

	fcm "github.com/andydunstall/figg/fcm/sdk"
	figg "github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestReconnect_ReconnectAfterConnectionDrops(t *testing.T) {
	fcmClient := fcm.NewFCM()
	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	// Add buffered chan to avoid blocking.
	connStateChan := make(chan figg.ConnState, 1)

	client, err := figg.Connect(
		cluster.Nodes[0].ProxyAddr,
		figg.WithLogger(setupLogger()),
		figg.WithConnStateChangeCB(func(s figg.ConnState) {
			connStateChan <- s
		}),
	)
	assert.Nil(t, err)
	defer client.Close()

	assert.Equal(t, figg.CONNECTED, <-connStateChan)

	// Disable the networking for the node and reenable after 2 second.
	fcmClient.AddChaosPartition(cluster.Nodes[0].ID, fcm.ChaosConfig{
		Duration: 2,
	})

	// The client should detect its disconnected then reconnect.
	assert.Equal(t, figg.DISCONNECTED, <-connStateChan)
	assert.Equal(t, figg.CONNECTED, <-connStateChan)
}
