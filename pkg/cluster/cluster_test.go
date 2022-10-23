package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCluster_Locate(t *testing.T) {
	c := NewCluster("peer-local")
	for i := 0; i != 10; i++ {
		c.NotifyJoin(fmt.Sprintf("peer-%d", i))
	}

	assert.Equal(t, "peer-3", c.Locate("foo"))
	assert.Equal(t, "peer-1", c.Locate("bar"))
	assert.Equal(t, "peer-2", c.Locate("car"))
}
