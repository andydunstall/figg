package server

import (
	"net"

	"github.com/andydunstall/figg/service/pkg/topic"
	"github.com/andydunstall/figg/utils"
	"go.uber.org/zap"
)

type Server struct {
	broker *topic.Broker
	logger *zap.Logger
}

func NewServer(broker *topic.Broker, logger *zap.Logger) *Server {
	s := &Server{
		broker: broker,
		logger: logger,
	}
	return s
}

func (s *Server) stream(c net.Conn) {
	addr := c.RemoteAddr().String()
	s.logger.Debug(
		"client connected",
		zap.String("addr", addr),
	)

	client := NewClient(utils.NewTCPConnection(c), s.broker)
	defer client.Shutdown()

	client.Serve()
}

func (s *Server) Serve(lis net.Listener) error {
	for {
		c, err := lis.Accept()
		if err != nil {
			return err
		}
		go s.stream(c)
	}
}
