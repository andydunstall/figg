package cluster

import (
	"sync"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Cluster struct {
	ID              string
	nodes           map[string]*Node
	portAllocator   *PortAllocator
	toxiproxyClient *toxiproxy.Client

	logger *zap.Logger

	mu sync.Mutex
}

func NewCluster(portAllocator *PortAllocator, toxiproxyClient *toxiproxy.Client, logger *zap.Logger) *Cluster {
	return &Cluster{
		ID:              uuid.New().String(),
		nodes:           make(map[string]*Node),
		portAllocator:   portAllocator,
		toxiproxyClient: toxiproxyClient,
		logger:          logger,
		mu:              sync.Mutex{},
	}
}

func (c *Cluster) AddNode() (*Node, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, err := NewNode(c.portAllocator, c.toxiproxyClient)
	if err != nil {
		return nil, err
	}
	c.nodes[node.ID] = node

	return node, nil

}

func (c *Cluster) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.nodes {
		c.logger.Debug("shutting down node", zap.String("node-id", node.ID))
		if err := node.Shutdown(); err != nil {
			c.logger.Warn("failed to shutdown node", zap.String("node-id", node.ID), zap.Error(err))
		}
	}

	return nil
}
