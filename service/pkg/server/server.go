package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/andydunstall/wombat/service/pkg/conn"
	"github.com/andydunstall/wombat/service/pkg/topic"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	router   *mux.Router
	broker   *topic.Broker
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
		broker:   topic.NewBroker(),
		upgrader: upgrader,
		logger:   logger,
	}
	s.addRoutes()
	return s
}

func (s *Server) addRoutes() {
	s.router.HandleFunc("/v1/ws", s.wsStream).Methods(http.MethodGet)
}

func (s *Server) wsStream(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Debug("failed to upgrade connection", zap.Error(err))
		return
	}
	defer ws.Close()

	addr := ws.RemoteAddr().String()
	s.logger.Debug(
		"ws stream connected",
		// zap.String("topic", topicName),
		zap.String("addr", addr),
		zap.String("offset", r.URL.Query().Get("offset")),
	)

	t := s.broker.GetTopic("tmp")
	transport := conn.NewWSTransport(ws)
	c := conn.NewProtocolConnection(transport)
	defer c.Close()

	var sub *conn.Subscription
	if r.URL.Query().Get("offset") != "" {
		offset, err := strconv.ParseUint(r.URL.Query().Get("offset"), 10, 64)
		if err != nil {
			s.logger.Debug("invalid offset param", zap.Error(err))
			// Fall back to the latest message if the offset is invalid.
			sub = conn.NewSubscription(t, c)
		} else {
			sub = conn.NewSubscriptionWithOffset(t, c, offset)
		}
	} else {
		sub = conn.NewSubscription(t, c)
	}
	defer sub.Shutdown()

	for {
		m, err := c.Recv()
		if err != nil {
			s.logger.Debug("failed to read from connection", zap.Error(err))
			return
		}
		switch m.Type {
		case conn.TypePing:
			c.Send(conn.NewPongMessage(m.Ping.Timestamp))
		case conn.TypeAttach:
			c.Send(conn.NewAttachedMessage())
		case conn.TypePublish:
			c.Send(conn.NewPayloadMessage(
				m.Publish.Topic,
				0,
				m.Publish.Payload,
			))
		}
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
