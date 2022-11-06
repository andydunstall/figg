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

func (c *Cluster) GetNode(id string) (*Node, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.nodes[id]
	return node, ok
}

func (c *Cluster) AddNode() (*Node, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, err := NewNode(c.portAllocator, c.toxiproxyClient, c.logger)
	if err != nil {
		return nil, err
	}
	c.nodes[node.ID] = node

	c.logger.Debug(
		"node added",
		zap.String("cluster-id", c.ID),
		zap.String("proxy-addr", node.Addr),
		zap.String("node-id", node.ID),
	)

	return node, nil
}

func (c *Cluster) RemoveNode(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.nodes[id]
	if !ok {
		// If not found do nothing.
		return nil
	}

	if err := node.Shutdown(); err != nil {
		return err
	}
	delete(c.nodes, id)

	c.logger.Debug(
		"node removed",
		zap.String("cluster-id", c.ID),
		zap.String("node-id", node.ID),
	)

	return nil
}

func (c *Cluster) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.nodes {
		if err := node.Shutdown(); err != nil {
			c.logger.Warn("failed to shutdown node", zap.String("node-id", node.ID), zap.Error(err))
		}
	}

	c.logger.Debug("cluster shutdown", zap.String("cluster-id", c.ID))

	return nil
}
