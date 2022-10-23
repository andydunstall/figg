package cluster

import (
	"sort"
	"sync"

	"github.com/spaolacci/murmur3"
	"go.uber.org/zap"
)

// Cluster tracks the active peers in the cluster the locates the correct peer
// for each topic.
type Cluster struct {
	peers map[string]interface{}

	logger *zap.Logger
	mu     sync.RWMutex
}

// NewCluster returns a new cluster with only the given peer as a member.
func NewCluster(peerID string, logger *zap.Logger) *Cluster {
	return &Cluster{
		peers: map[string]interface{}{
			peerID: struct{}{},
		},
		logger: logger,
		mu:     sync.RWMutex{},
	}
}

// Locate returns the peer ID the given topic name is located.
//
// The peers in the cluster form a hash ring, where the topic is located at
// the next peer walking around the ring.
func (c *Cluster) Locate(name string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Currently just using an inefficient lookup - can optimise later as
	// needed.

	peersByHash := map[uint64]string{}
	peerHashes := []uint64{}
	for peer, _ := range c.peers {
		peerHash := murmur3.Sum64WithSeed([]byte(peer), 0)
		peersByHash[peerHash] = peer
		peerHashes = append(peerHashes, peerHash)
	}
	sort.Slice(peerHashes, func(i, j int) bool {
		return peerHashes[i] < peerHashes[j]
	})

	nameHash := murmur3.Sum64WithSeed([]byte(name), 0)
	for _, peerHash := range peerHashes {
		if nameHash <= peerHash {
			return peersByHash[peerHash]
		}
	}
	// Wrap around the ring to the first hash if at the end of the ring.
	return peersByHash[peerHashes[0]]
}

// NotifyJoin adds a new peer to the hash ring.
func (c *Cluster) NotifyJoin(peerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("cluster: peer added", zap.String("peer", peerID))

	c.peers[peerID] = struct{}{}
}
