package wombat

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/andydunstall/scuttlebutt"
	"github.com/andydunstall/wombat/service/pkg/cluster"
	"github.com/andydunstall/wombat/service/pkg/config"
	"github.com/andydunstall/wombat/service/pkg/server"
	"go.uber.org/zap"
)

func setupLogger(debugMode bool) (*zap.Logger, error) {
	if debugMode {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		return logger, nil
	}
	return zap.NewProduction()
}

func Run(config config.Config) {
	logger, err := setupLogger(config.Verbose)
	if err != nil {
		log.Fatalf("failed to setup logger: %s", err)
	}
	defer logger.Sync()

	logger.Info("starting wombat")

	cluster := cluster.NewCluster(config.GossipPeerID, logger)
	gossip, err := scuttlebutt.Create(&scuttlebutt.Config{
		ID:             config.GossipPeerID,
		BindAddr:       config.GossipAddr,
		NodeSubscriber: cluster,
		Logger:         logger.With(zap.String("tag", "gossip")),
	})
	if err != nil {
		logger.Fatal("failed to setup gossip", zap.Error(err))
	}
	defer gossip.Shutdown()

	if config.GossipSeeds != "" {
		gossip.Seed(strings.Split(config.GossipSeeds, ","))
	}

	server := server.NewServer(logger)

	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		logger.Fatal("failed to start listener", zap.Error(err))
	}

	go func() {
		if err := server.Serve(lis); err != nil {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c

	logger.Info("starting shut down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	err = server.Shutdown(ctx)
	logger.Info("finished shut down", zap.Error(err))
}
