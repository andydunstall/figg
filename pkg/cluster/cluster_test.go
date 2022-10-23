package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCluster_Locate(t *testing.T) {
	c := NewCluster("peer-local", zap.NewNop())
	for i := 0; i != 10; i++ {
		c.NotifyJoin(fmt.Sprintf("peer-%d", i))
	}

	assert.Equal(t, "peer-3", c.Locate("foo"))
	assert.Equal(t, "peer-1", c.Locate("bar"))
	assert.Equal(t, "peer-2", c.Locate("car"))
}
