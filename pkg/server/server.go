package server

import (
	"net/http"

	"github.com/andydunstall/wombat/pkg/broker"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	router   *mux.Router
	broker   *broker.Broker
	upgrader websocket.Upgrader
	logger   *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	router := mux.NewRouter()

	upgrader := websocket.Upgrader{}
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
	s.router.HandleFunc("/v1/ws/{topic}", s.wsStream).Methods(http.MethodGet)
}

func (s *Server) wsStream(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	topicName := reqVars["topic"]

	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Debug("failed to upgrade connection", zap.Error(err))
		return
	}
	defer c.Close()

	addr := c.RemoteAddr().String()
	s.logger.Debug("ws stream connected", zap.String("addr", addr))

	topic := s.broker.GetTopic(topicName)
	topic.Subscribe(addr, c)
	defer topic.Unsubscribe(addr)

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			s.logger.Debug("failed to read from ws connection", zap.Error(err))
			return
		}
		topic.Publish(mt, message)
	}
}

func (s *Server) Listen(addr string) error {
	s.logger.Info(
		"starting server",
		zap.String("addr", addr),
	)

	http.Handle("/", s.router)
	return http.ListenAndServe(addr, nil)
}
