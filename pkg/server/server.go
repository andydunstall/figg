package server

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/andydunstall/wombat/pkg/broker"
	"github.com/andydunstall/wombat/pkg/topic"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	router   *mux.Router
	broker   *broker.Broker
	upgrader websocket.Upgrader
	srv      *http.Server
	logger   *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	router := mux.NewRouter()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	s := &Server{
		router:   router,
		broker:   broker.NewBroker(),
		upgrader: upgrader,
		logger:   logger,
	}
	s.addRoutes()
	return s
}

func (s *Server) addRoutes() {
	s.router.HandleFunc("/v1/{topic}", s.restPublish).Methods(http.MethodPost)
	s.router.HandleFunc("/v1/{topic}/ws", s.wsStream).Methods(http.MethodGet)
}

func (s *Server) restPublish(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	topicName := reqVars["topic"]

	if r.Body == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	addr := r.RemoteAddr
	s.logger.Debug(
		"rest publish",
		zap.String("topic", topicName),
		zap.String("addr", addr),
	)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	b := buf.Bytes()

	t := s.broker.GetTopic(topicName)
	t.Publish(b)
}

func (s *Server) wsStream(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	topicName := reqVars["topic"]

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Debug("failed to upgrade connection", zap.Error(err))
		return
	}
	defer ws.Close()

	addr := ws.RemoteAddr().String()
	s.logger.Debug(
		"ws stream connected",
		zap.String("topic", topicName),
		zap.String("addr", addr),
		zap.String("offset", r.URL.Query().Get("offset")),
	)

	t := s.broker.GetTopic(topicName)
	conn := NewWSConn(ws)
	var sub *topic.Subscription
	if r.URL.Query().Get("offset") != "" {
		offset, err := strconv.ParseUint(r.URL.Query().Get("offset"), 10, 64)
		if err != nil {
			s.logger.Debug("invalid offset param", zap.Error(err))
			// Fall back to the latest message if the offset is invalid.
			sub = topic.NewSubscription(t, conn)
		} else {
			sub = topic.NewSubscriptionWithOffset(t, conn, offset)
		}
	} else {
		sub = topic.NewSubscription(t, conn)
	}
	defer sub.Shutdown()

	for {
		b, err := conn.Recv()
		if err != nil {
			s.logger.Debug("failed to read from connection", zap.Error(err))
			return
		}
		t.Publish(b)
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
