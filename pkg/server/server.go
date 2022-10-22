package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	router   *mux.Router
	upgrader websocket.Upgrader
	logger   *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	router := mux.NewRouter()

	upgrader := websocket.Upgrader{}
	s := &Server{
		router:   router,
		upgrader: upgrader,
		logger:   logger,
	}
	s.addRoutes()
	return s
}

func (s *Server) addRoutes() {
	s.router.HandleFunc("/v1/ws/stream", s.wsStream).Methods(http.MethodGet)
}

func (s *Server) wsStream(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Debug("failed to upgrade connection", zap.Error(err))
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			s.logger.Debug("failed to read from ws connection", zap.Error(err))
			break
		}
		if err = c.WriteMessage(mt, message); err != nil {
			s.logger.Debug("failed to write to ws connection", zap.Error(err))
			break
		}
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
