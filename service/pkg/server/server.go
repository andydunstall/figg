package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/andydunstall/figg/service/pkg/conn"
	"github.com/andydunstall/figg/service/pkg/topic"
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
		zap.String("addr", addr),
	)

	transport := conn.NewWSTransport(ws)
	c := conn.NewProtocolConnection(transport)
	defer c.Close()

	subscriptions := topic.NewSubscriptions(s.broker)
	defer subscriptions.Shutdown()

	doneCh := make(chan interface{})

	go func() {
		for {
			select {
			case m := <-subscriptions.MessageCh():
				c.Send(conn.NewPayloadMessage(m.Topic, m.Offset, m.Message))
			case <-doneCh:
				return
			}
		}
	}()

	for {
		m, err := c.Recv()
		if err != nil {
			s.logger.Debug("failed to read from connection", zap.Error(err))
			break
		}
		switch m.Type {
		case conn.TypePing:
			c.Send(conn.NewPongMessage(m.Ping.Timestamp))
		case conn.TypeAttach:
			if m.Attach.Offset != "" {
				offset, err := strconv.ParseUint(m.Attach.Offset, 10, 64)
				if err != nil {
					// If the offset is invalid subscribe without.
					subscriptions.AddSubscription(m.Attach.Topic)
				} else {
					subscriptions.AddSubscriptionFromOffset(m.Attach.Topic, offset)
				}
			} else {
				subscriptions.AddSubscription(m.Attach.Topic)
			}
			c.Send(conn.NewAttachedMessage())
		case conn.TypePublish:
			topic := s.broker.GetTopic(m.Publish.Topic)
			topic.Publish(m.Publish.Payload)
			c.Send(conn.NewACKMessage(m.Publish.SeqNum))
		}
	}

	close(doneCh)
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
