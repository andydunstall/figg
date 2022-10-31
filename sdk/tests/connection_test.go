package sdk

import (
	"context"
	"testing"
	"time"

	"github.com/andydunstall/wombat/sdk"
	"github.com/andydunstall/wombat/wcm/sdk"
	"github.com/stretchr/testify/assert"
)

func TestConnection_Connect(t *testing.T) {
	wcm, err := wcm.Connect()
	assert.Nil(t, err)
	defer wcm.Close()

	cluster, err := wcm.CreateCluster(context.Background())
	assert.Nil(t, err)
	defer cluster.Close(context.Background())

	node, err := cluster.AddNode(context.Background())
	assert.Nil(t, err)

	// TODO(AD) Shouldn't need to wait. SDK should retry.
	<-time.After(time.Second)

	wombat, err := wombat.NewWombat(node.Addr, "foo")
	assert.Nil(t, err)
	defer wombat.Shutdown()
}
