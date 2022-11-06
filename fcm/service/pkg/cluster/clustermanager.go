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
	portAllocator   *PortAllocator
	toxiproxyClient *toxiproxy.Client

	mu sync.Mutex

	logger *zap.Logger
}

func NewClusterManager(logger *zap.Logger) *ClusterManager {
	return &ClusterManager{
		clusters:        make(map[string]*Cluster),
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

func (m *ClusterManager) Add() *Cluster {
	m.mu.Lock()
	defer m.mu.Unlock()

	cluster := NewCluster(m.portAllocator, m.toxiproxyClient, m.logger)
	m.clusters[cluster.ID] = cluster

	m.logger.Info("added cluster", zap.String("cluster-id", cluster.ID))

	return cluster
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
