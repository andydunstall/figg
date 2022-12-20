package server

import (
	"net"

	"github.com/andydunstall/figg/service/pkg/topic"
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

func (s *Server) Serve(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go s.stream(NewConnection(conn, s.broker), conn.RemoteAddr().String())
	}
}

func (s *Server) stream(conn *Connection, addr string) {
	defer conn.Close()

	s.logger.Debug(
		"client connected",
		zap.String("addr", addr),
	)

	for {
		if err := conn.Recv(); err != nil {
			s.logger.Debug("client connection closed")
			return
		}
	}
}
