package service

import (
	"net"
	"sync"

	"github.com/andydunstall/figg/server/pkg/admin/server"
	"github.com/andydunstall/figg/server/pkg/config"
	"go.uber.org/zap"
)

type AdminService struct {
	config config.Config
	logger *zap.Logger
	lis    net.Listener
	wg     sync.WaitGroup
}

func NewAdminService(config config.Config, logger *zap.Logger) *AdminService {
	return &AdminService{
		config: config,
		logger: logger,
		wg:     sync.WaitGroup{},
	}
}

// Serve starts the service and returns the server address.
//
// Note this won't always be the same as the configured address, such as if
// port 0 used the system will assign a free port.
func (s *AdminService) Serve() (string, error) {
	s.logger.Info("starting admin service")

	server := server.NewServer()

	lis, err := net.Listen("tcp", s.config.AdminAddr)
	if err != nil {
		return "", err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		server.Serve(lis)
	}()

	s.lis = lis
	return lis.Addr().String(), nil
}

// Close stops the server and wait for them to exit.
func (s *AdminService) Close() {
	// Close the listener which will cause the server goroutine to exit.
	s.lis.Close()
	s.wg.Wait()
}
