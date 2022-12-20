package sdk

import (
	"testing"

	fcm "github.com/andydunstall/figg/fcm/sdk"
	figg "github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestConnect_ConnectThenClose(t *testing.T) {
	fcmClient := fcm.NewFCM()
	cluster, err := fcmClient.AddCluster()
	assert.Nil(t, err)
	defer fcmClient.RemoveCluster(cluster.ID)

	client, err := figg.Connect(
		cluster.Nodes[0].Addr,
		figg.WithLogger(setupLogger()),
	)
	assert.Nil(t, err)
	defer client.Close()
}

func TestConnect_ServerUnreachable(t *testing.T) {
	_, err := figg.Connect("1.2.3.4:8119", figg.WithLogger(setupLogger()))
	assert.Error(t, err)
}
