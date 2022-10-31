package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/andydunstall/wombat/wcm/service/pkg/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type clusterInfo struct {
	ID string `json:"id,omitempty"`
}

type nodeInfo struct {
	ID   string `json:"id,omitempty"`
	Addr string `json:"addr,omitempty"`
}

type Server struct {
	clusterManager *cluster.ClusterManager

	router *mux.Router
	srv    *http.Server

	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	router := mux.NewRouter()

	s := &Server{
		clusterManager: cluster.NewClusterManager(logger),
		router:         router,
		srv:            nil,
		logger:         logger,
	}
	s.addRoutes()
	return s
}

func (s *Server) addRoutes() {
	s.router.HandleFunc("/v1/clusters", s.addCluster).Methods(http.MethodPost)
	s.router.HandleFunc("/v1/clusters/{clusterID}", s.removeCluster).Methods(http.MethodDelete)

	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes", s.addNode).Methods(http.MethodPost)
	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes/{nodeID}", s.removeNode).Methods(http.MethodDelete)
}

func (s *Server) addCluster(w http.ResponseWriter, r *http.Request) {
	cluster := s.clusterManager.Add()

	resp := clusterInfo{
		ID: cluster.ID,
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		s.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) removeCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("remove cluster: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.clusterManager.Remove(clusterID)
}

func (s *Server) addNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("add node: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	cluster, ok := s.clusterManager.Get(clusterID)
	if !ok {
		s.logger.Debug("add node: cluster node found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	node, err := cluster.AddNode()
	if err != nil {
		s.logger.Error("add node: failed to add node", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := nodeInfo{
		ID:   node.ID,
		Addr: node.Addr,
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		s.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) removeNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("remove node: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	nodeID := vars["nodeID"]
	if nodeID == "" {
		s.logger.Debug("remove node: missing node ID")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cluster, ok := s.clusterManager.Get(clusterID)
	if !ok {
		s.logger.Debug("remove node: cluster node found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := cluster.RemoveNode(nodeID); err != nil {
		s.logger.Error("remove node: failed to remove node", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) Serve(lis net.Listener) error {
	srv := &http.Server{
		Handler:      s.router,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	s.srv = srv
	return srv.Serve(lis)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}
