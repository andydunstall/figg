package cluster

import (
	"sync"

	fcm "github.com/andydunstall/figg/fcm/lib"
	"go.uber.org/zap"
)

type ClusterManager struct {
	clusters map[string]*Cluster
	nodes    map[string]*fcm.Node

	mu sync.Mutex

	logger *zap.Logger
}

func NewClusterManager(logger *zap.Logger) *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*Cluster),
		nodes:    make(map[string]*fcm.Node),
		mu:       sync.Mutex{},
		logger:   logger,
	}
}

func (m *ClusterManager) Get(id string) (*Cluster, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, ok := m.clusters[id]
	return cluster, ok
}

func (m *ClusterManager) GetNode(id string) (*fcm.Node, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	return node, ok
}

func (m *ClusterManager) Add() (*Cluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, err := NewCluster(m.logger)
	if err != nil {
		return nil, err
	}

	m.clusters[cluster.ID] = cluster
	for _, node := range cluster.Nodes {
		m.nodes[node.ID] = node
	}

	m.logger.Info("added cluster", zap.String("cluster-id", cluster.ID))

	return cluster, nil
}

func (m *ClusterManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, ok := m.clusters[id]
	if !ok {
		// If not found do nothing.
		return
	}

	cluster.Shutdown()
	delete(m.clusters, id)

	m.logger.Info("removed cluster", zap.String("cluster-id", cluster.ID))
}
