package tests

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/andydunstall/wombat/pkg/server"
	"go.uber.org/zap"
)

type Server struct {
	wombat *server.Server
	wg     sync.WaitGroup
}

func NewServer() *Server {
	return &Server{
		wombat: server.NewServer(zap.NewNop()),
		wg:     sync.WaitGroup{},
	}
}

func (s *Server) Run() (string, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.wombat.Serve(lis)
	}()

	return lis.Addr().String(), nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return s.wombat.Shutdown(ctx)
}
