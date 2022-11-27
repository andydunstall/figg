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
	ID        string
	Addr      string
	ProxyAddr string
	proxy     *toxiproxy.Proxy

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
		ID:        id,
		Addr:      listenAddr,
		ProxyAddr: proxyAddr,
		proxy:     proxy,
		logger:    logger,
		doneCh:    doneCh,
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

func (n *Node) Partition(duration int, repeat int) {
	fmt.Println("partition", duration, repeat)

	n.PartitionFor(duration)
	if repeat != 0 {
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(repeat))
			for {
				select {
				case <-ticker.C:
					n.PartitionFor(duration)
				}
			}
		}()
	}
}

func (n *Node) PartitionFor(duration int) {
	fmt.Println("partition for", duration)
	n.Disable()
	if duration != 0 {
		go func() {
			<-time.After(time.Duration(duration) * time.Second)
			n.Enable()
		}()
	}
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
