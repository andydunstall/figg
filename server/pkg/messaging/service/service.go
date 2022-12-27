package service

import (
	"net"
	"sync"

	"github.com/andydunstall/figg/server/pkg/config"
	"github.com/andydunstall/figg/server/pkg/messaging/server"
	"github.com/andydunstall/figg/server/pkg/topic"
	"go.uber.org/zap"
)

// MessagingService implements the core pub/sub service.
type MessagingService struct {
	config config.Config
	logger *zap.Logger
	// lis is the services network listener. nil if the service is not
	// running.
	lis net.Listener
	wg  sync.WaitGroup
}

func NewMessagingService(config config.Config, logger *zap.Logger) *MessagingService {
	return &MessagingService{
		config: config,
		logger: logger,
		wg:     sync.WaitGroup{},
	}
}

// Serve starts the service and returns the server address.
//
// Note this won't always be the same as the configured address, such as if
// port 0 used the system will assign a free port.
func (s *MessagingService) Serve() (string, error) {
	s.logger.Info("starting messaging service")

	server := server.NewServer(topic.NewBroker(topic.Options{
		Persisted:   !s.config.CommitLogInMemory,
		Dir:         s.config.CommitLogDir,
		SegmentSize: s.config.CommitLogSegmentSize,
	}), s.logger)

	lis, err := net.Listen("tcp", s.config.Addr)
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
func (s *MessagingService) Close() {
	// Close the listener which will cause the server goroutine to exit.
	s.lis.Close()
	s.wg.Wait()
}
