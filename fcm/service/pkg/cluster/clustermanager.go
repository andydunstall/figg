package cluster

import (
	"sync"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"go.uber.org/zap"
)

const (
	// Using ports that won't be used by the system.
	PortRangeFrom = 40000
	PortRangeTo   = 60000
)

type ClusterManager struct {
	clusters        map[string]*Cluster
	nodes           map[string]*Node
	portAllocator   *PortAllocator
	toxiproxyClient *toxiproxy.Client

	mu sync.Mutex

	logger *zap.Logger
}

func NewClusterManager(logger *zap.Logger) *ClusterManager {
	return &ClusterManager{
		clusters:        make(map[string]*Cluster),
		nodes:           make(map[string]*Node),
		portAllocator:   NewPortAllocator(PortRangeFrom, PortRangeTo),
		toxiproxyClient: toxiproxy.NewClient("localhost:8474"),
		mu:              sync.Mutex{},
		logger:          logger,
	}
}

func (m *ClusterManager) Get(id string) (*Cluster, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, ok := m.clusters[id]
	return cluster, ok
}

func (m *ClusterManager) GetNode(id string) (*Node, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	return node, ok
}

func (m *ClusterManager) Add() (*Cluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster, err := NewCluster(m.portAllocator, m.toxiproxyClient, m.logger)
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
