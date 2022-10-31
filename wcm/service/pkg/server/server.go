package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/andydunstall/wombat/wcm/service/pkg/cluster"
	pb "github.com/andydunstall/wombat/wcm/service/pkg/rpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	PortRangeFrom = 10000
	PortRangeTo   = 20000
)

type Server struct {
	clusters      map[string]*cluster.Cluster
	portAllocator *cluster.PortAllocator

	mu sync.Mutex

	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{
		clusters:      make(map[string]*cluster.Cluster),
		portAllocator: cluster.NewPortAllocator(PortRangeFrom, PortRangeTo),
		mu:            sync.Mutex{},
		logger:        logger,
	}
}

func (s *Server) CreateCluster(ctx context.Context, req *pb.EmptyMessage) (*pb.ClusterInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cluster := cluster.NewCluster(s.portAllocator, s.logger)
	s.clusters[cluster.ID] = cluster

	s.logger.Info("cluster created", zap.String("id", cluster.ID))

	return &pb.ClusterInfo{
		Id: cluster.ID,
	}, nil
}

func (s *Server) ClusterAddNode(ctx context.Context, req *pb.ClusterInfo) (*pb.NodeInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cluster, ok := s.clusters[req.Id]
	if !ok {
		s.logger.Info(
			"requested cluster not found",
			zap.String("cluster-id", req.Id),
		)
		return nil, fmt.Errorf("cluster not found")
	}
	node, err := cluster.AddNode()
	if err != nil {
		return nil, err
	}

	s.logger.Info(
		"node added",
		zap.String("cluster-id", cluster.ID),
		zap.String("node-id", node.ID),
		zap.String("node-addr", node.Addr),
	)

	return &pb.NodeInfo{
		Id:   node.ID,
		Addr: node.Addr,
	}, nil
}

func (s *Server) ClusterClose(ctx context.Context, req *pb.ClusterInfo) (*pb.EmptyMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cluster, ok := s.clusters[req.Id]
	if !ok {
		s.logger.Info("cluster not found", zap.String("cluster-id", req.Id))
		return nil, fmt.Errorf("cluster not found")
	}

	cluster.Shutdown()
	delete(s.clusters, req.Id)

	s.logger.Info("cluster shutdown", zap.String("id", cluster.ID))

	return &pb.EmptyMessage{}, nil
}

func (s *Server) Listen(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.logger.Info("starting wcm service", zap.String("addr", addr))

	grpcServer := grpc.NewServer()
	pb.RegisterWCMServer(grpcServer, s)
	return grpcServer.Serve(lis)
}
