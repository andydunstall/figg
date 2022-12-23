package sdk

import (
	"testing"

	fcm "github.com/andydunstall/figg/fcm/lib"
	figg "github.com/andydunstall/figg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestConnect_ConnectThenClose(t *testing.T) {
	node, err := fcm.NewNode(setupLogger())
	assert.Nil(t, err)
	defer node.Shutdown()

	client, err := figg.Connect(
		node.Addr,
		figg.WithLogger(setupLogger()),
	)
	assert.Nil(t, err)
	defer client.Close()
}

func TestConnect_ServerUnreachable(t *testing.T) {
	_, err := figg.Connect("1.2.3.4:8119", figg.WithLogger(setupLogger()))
	assert.Error(t, err)
}
