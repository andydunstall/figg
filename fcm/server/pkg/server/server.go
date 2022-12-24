package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/andydunstall/figg/fcm/server/pkg/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type clusterInfo struct {
	ID    string     `json:"id,omitempty"`
	Nodes []nodeInfo `json:"nodes,omitempty"`
}

type nodeInfo struct {
	ID        string `json:"id,omitempty"`
	Addr      string `json:"addr,omitempty"`
	ProxyAddr string `json:"proxy_addr,omitempty"`
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
	s.router.Use(s.requestLogger)

	s.router.HandleFunc("/v1/clusters", s.addCluster).Methods(http.MethodPost)
	s.router.HandleFunc("/v1/clusters/{clusterID}", s.removeCluster).Methods(http.MethodDelete)

	s.router.HandleFunc("/v1/chaos/partition/{nodeID}", s.chaosPartition).Methods(http.MethodPost)
}

func (s *Server) addCluster(w http.ResponseWriter, r *http.Request) {
	cluster, err := s.clusterManager.Add()
	if err != nil {
		s.logger.Error("add cluster: failed to add cluster", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := clusterInfo{
		ID:    cluster.ID,
		Nodes: []nodeInfo{},
	}
	for _, node := range cluster.Nodes {
		resp.Nodes = append(resp.Nodes, nodeInfo{
			ID:        node.ID,
			Addr:      node.Addr,
			ProxyAddr: node.ProxyAddr,
		})
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

func (s *Server) chaosPartition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeID := vars["nodeID"]
	if nodeID == "" {
		s.logger.Debug("chaos partition: missing node ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	node, ok := s.clusterManager.GetNode(nodeID)
	if !ok {
		s.logger.Debug("chaos partition: node not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	duration := 0
	durationStr := r.URL.Query().Get("duration")
	if durationStr != "" {
		var err error
		duration, err = strconv.Atoi(durationStr)
		if err != nil {
			s.logger.Debug("chaos partition: invalid duration", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
	}

	repeat := 0
	repeatStr := r.URL.Query().Get("repeat")
	if repeatStr != "" {
		var err error
		repeat, err = strconv.Atoi(repeatStr)
		if err != nil {
			s.logger.Debug("chaos partition: invalid repeat", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
	}

	node.Partition(duration, repeat)
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

func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug(
			"request",
			zap.String("url", r.URL.String()),
			zap.String("method", r.Method),
		)
		next.ServeHTTP(w, r)
	})
}
