package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/andydunstall/figg/fcm/service/pkg/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type clusterInfo struct {
	ID string `json:"id,omitempty"`
}

type nodeInfo struct {
	ID        string `json:"id,omitempty"`
	Addr      string `json:"addr,omitempty"`
	ProxyAddr string `json:"proxy_addr,omitempty"`
}

type scenarioInfo struct {
	ID string `json:"id,omitempty"`
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

	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes", s.addNode).Methods(http.MethodPost)
	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes/{nodeID}", s.removeNode).Methods(http.MethodDelete)

	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes/{nodeID}/enable", s.enableNode).Methods(http.MethodPost)
	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes/{nodeID}/disable", s.disableNode).Methods(http.MethodPost)

	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes/{nodeID}/latency", s.addLatency).Methods(http.MethodPost)

	s.router.HandleFunc("/v1/clusters/{clusterID}/nodes/{nodeID}/chaos/{scenarioID}", s.removeChaosScenario).Methods(http.MethodDelete)
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
		ID:        node.ID,
		Addr:      node.Addr,
		ProxyAddr: node.ProxyAddr,
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

func (s *Server) enableNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("enable node: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	nodeID := vars["nodeID"]
	if nodeID == "" {
		s.logger.Debug("enable node: missing node ID")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cluster, ok := s.clusterManager.Get(clusterID)
	if !ok {
		s.logger.Debug("enable node: cluster not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	node, ok := cluster.GetNode(nodeID)
	if !ok {
		s.logger.Debug("enable node: node not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := node.Enable(); err != nil {
		s.logger.Error("enable node: failed to connect node", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) addLatency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("add latency: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	nodeID := vars["nodeID"]
	if nodeID == "" {
		s.logger.Debug("add latency: missing node ID")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	latencyStr := r.URL.Query().Get("latency")
	if latencyStr == "" {
		// Default to 1s of latency.
		latencyStr = "1000"
	}

	latencyMS, err := strconv.Atoi(latencyStr)
	if err != nil {
		s.logger.Debug("add latency: invalid latency")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	latency := time.Duration(latencyMS) * time.Millisecond

	cluster, ok := s.clusterManager.Get(clusterID)
	if !ok {
		s.logger.Debug("add latency: cluster not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	node, ok := cluster.GetNode(nodeID)
	if !ok {
		s.logger.Debug("add latency: node not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id, err := node.AddLatency(latency)
	if err != nil {
		s.logger.Error("add latency: failed to connect node", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := scenarioInfo{
		ID: id,
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		s.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) removeChaosScenario(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("add latency: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	nodeID := vars["nodeID"]
	if nodeID == "" {
		s.logger.Debug("add latency: missing node ID")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	scenarioID := vars["scenarioID"]
	if scenarioID == "" {
		s.logger.Debug("add latency: missing scenario ID")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cluster, ok := s.clusterManager.Get(clusterID)
	if !ok {
		s.logger.Debug("add latency: cluster not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	node, ok := cluster.GetNode(nodeID)
	if !ok {
		s.logger.Debug("add latency: node not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := node.RemoveScenario(scenarioID); err != nil {
		s.logger.Error("add latency: failed to remove scenario", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) disableNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["clusterID"]
	if clusterID == "" {
		s.logger.Debug("disable node: missing cluster ID")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	nodeID := vars["nodeID"]
	if nodeID == "" {
		s.logger.Debug("disable node: missing node ID")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cluster, ok := s.clusterManager.Get(clusterID)
	if !ok {
		s.logger.Debug("disable node: cluster not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	node, ok := cluster.GetNode(nodeID)
	if !ok {
		s.logger.Debug("disable node: node not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := node.Disable(); err != nil {
		s.logger.Error("disable node: failed to disconnect node", zap.Error(err))
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
