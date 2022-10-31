package cluster

import (
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Cluster struct {
	ID            string
	nodes         map[string]*Node
	portAllocator *PortAllocator

	logger *zap.Logger

	mu sync.Mutex
}

func NewCluster(portAllocator *PortAllocator, logger *zap.Logger) *Cluster {
	return &Cluster{
		ID:            uuid.New().String(),
		nodes:         make(map[string]*Node),
		portAllocator: portAllocator,
		logger:        logger,
		mu:            sync.Mutex{},
	}
}

func (c *Cluster) AddNode() (*Node, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, err := NewNode(c.portAllocator)
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
		c.logger.Debug("signalling node", zap.String("node-id", node.ID))
		if err := node.Kill(); err != nil {
			c.logger.Warn("failed to kill node", zap.String("node-id", node.ID), zap.Error(err))
		}
	}

	for _, node := range c.nodes {
		c.logger.Debug("waiting for node", zap.String("node-id", node.ID))
		if err := node.Wait(); err != nil {
			c.logger.Warn("failed to wait for node", zap.String("node-id", node.ID), zap.Error(err))
		}
	}

	return nil
}
