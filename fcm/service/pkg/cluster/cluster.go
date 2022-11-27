package cluster

import (
	"sync"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Cluster struct {
	ID              string
	Nodes           map[string]*Node
	portAllocator   *PortAllocator
	toxiproxyClient *toxiproxy.Client

	logger *zap.Logger

	mu sync.Mutex
}

func NewCluster(portAllocator *PortAllocator, toxiproxyClient *toxiproxy.Client, logger *zap.Logger) (*Cluster, error) {
	cluster := &Cluster{
		ID:              uuid.New().String(),
		Nodes:           make(map[string]*Node),
		portAllocator:   portAllocator,
		toxiproxyClient: toxiproxyClient,
		logger:          logger,
		mu:              sync.Mutex{},
	}
	_, err := cluster.AddNode()
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Cluster) AddNode() (*Node, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, err := NewNode(c.portAllocator, c.toxiproxyClient, c.logger)
	if err != nil {
		return nil, err
	}
	c.Nodes[node.ID] = node

	c.logger.Debug(
		"node added",
		zap.String("cluster-id", c.ID),
		zap.String("proxy-addr", node.Addr),
		zap.String("node-id", node.ID),
	)

	return node, nil
}

func (c *Cluster) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.Nodes {
		if err := node.Shutdown(); err != nil {
			c.logger.Warn("failed to shutdown node", zap.String("node-id", node.ID), zap.Error(err))
		}
	}

	c.logger.Debug("cluster shutdown", zap.String("cluster-id", c.ID))

	return nil
}
