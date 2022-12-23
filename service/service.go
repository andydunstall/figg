package service

import (
	"errors"
	"net"

	"github.com/andydunstall/figg/service/pkg/config"
	"github.com/andydunstall/figg/service/pkg/server"
	"github.com/andydunstall/figg/service/pkg/topic"
	"go.uber.org/zap"
)

type Figg struct {
	config config.Config
	server *server.Server
	logger *zap.Logger
	lis    net.Listener
}

func NewFigg(config config.Config, logger *zap.Logger) *Figg {
	server := server.NewServer(topic.NewBroker(topic.Options{
		Persisted:   !config.CommitLogInMemory,
		Dir:         config.CommitLogDir,
		SegmentSize: config.CommitLogSegmentSize,
	}), logger)
	return &Figg{
		config: config,
		server: server,
		logger: logger,
		lis:    nil,
	}
}

func (f *Figg) Serve() error {
	lis, err := net.Listen("tcp", f.config.Addr)
	if err != nil {
		f.logger.Fatal("failed to start listener", zap.Error(err))
	}
	return f.ServeWithListener(lis)
}

func (f *Figg) ServeWithListener(lis net.Listener) error {
	if f.lis != nil {
		return errors.New("already serving")
	}

	f.logger.Info("starting figg service", zap.Object("config", f.config))

	f.lis = lis
	return f.server.Serve(lis)
}

func (f *Figg) Close() error {
	if f.lis == nil {
		// Server not running.
		return nil
	}
	return f.lis.Close()
}
