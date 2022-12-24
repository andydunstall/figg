package cluster

import (
	"sync"

	fcm "github.com/andydunstall/figg/fcm/lib"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Cluster struct {
	ID    string
	Nodes map[string]*fcm.Node

	logger *zap.Logger

	mu sync.Mutex
}

func NewCluster(logger *zap.Logger) (*Cluster, error) {
	cluster := &Cluster{
		ID:     uuid.New().String()[:7],
		Nodes:  make(map[string]*fcm.Node),
		logger: logger,
		mu:     sync.Mutex{},
	}
	_, err := cluster.AddNode()
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Cluster) AddNode() (*fcm.Node, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, err := fcm.NewNode(c.logger)
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
