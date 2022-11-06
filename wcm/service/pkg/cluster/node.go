package cluster

import (
	"fmt"
	"os"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/andydunstall/figg/service"
	"github.com/andydunstall/figg/service/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Node struct {
	ID    string
	Addr  string
	proxy *toxiproxy.Proxy

	logger *zap.Logger

	doneCh chan interface{}
}

func NewNode(portAllocator *PortAllocator, toxiproxyClient *toxiproxy.Client, logger *zap.Logger) (*Node, error) {
	id := uuid.New().String()

	listenAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())
	proxyAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())

	proxyID := fmt.Sprintf("figg_%s", id)
	proxy, err := toxiproxyClient.CreateProxy(proxyID, proxyAddr, listenAddr)
	if err != nil {
		return nil, err
	}

	config := config.Config{
		Addr: listenAddr,
	}

	procLogger, err := newLogger(id)
	if err != nil {
		return nil, err
	}

	doneCh := make(chan interface{})

	go func() {
		service.Run(config, procLogger, doneCh)
	}()

	return &Node{
		ID:     id,
		Addr:   proxyAddr,
		proxy:  proxy,
		logger: logger,
		doneCh: doneCh,
	}, nil
}

func (n *Node) Enable() error {
	if err := n.proxy.Enable(); err != nil {
		n.logger.Error("failed to enable node", zap.String("node-id", n.ID), zap.Error(err))
	}

	n.logger.Debug("node enabled", zap.String("node-id", n.ID))
	return nil
}

func (n *Node) Disable() error {
	if err := n.proxy.Disable(); err != nil {
		n.logger.Error("failed to disable node", zap.String("node-id", n.ID), zap.Error(err))
	}

	n.logger.Debug("node disabled", zap.String("node-id", n.ID))
	return nil
}

func (n *Node) AddLatency(d time.Duration) (string, error) {
	id := uuid.New().String()
	_, err := n.proxy.AddToxic(id, "latency", "downstream", 1.0, toxiproxy.Attributes{
		"latency": d.Milliseconds(),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (n *Node) RemoveScenario(id string) error {
	return n.proxy.RemoveToxic(id)
}

func (n *Node) Shutdown() error {
	if err := n.proxy.Delete(); err != nil {
		return err
	}
	close(n.doneCh)
	return nil
}

func newLogger(id string) (*zap.Logger, error) {
	if err := createLogDir(id); err != nil {
		return nil, err
	}
	path := "out/" + id + "/out.log"

	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{path}
	return cfg.Build()
}

func createLogDir(id string) error {
	err := os.MkdirAll("out/"+id, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
