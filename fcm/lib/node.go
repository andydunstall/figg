package fcm

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/andydunstall/figg/service"
	"github.com/andydunstall/figg/service/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Node struct {
	ID        string
	Addr      string
	ProxyAddr string
	proxy     *Proxy

	logger *zap.Logger

	figg *service.Figg
}

func NewNode(logger *zap.Logger) (*Node, error) {
	id := uuid.New().String()[:7]

	// Create figg and proxy listeners, leaving the kernel to assign a free
	// port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	listenAddr := listener.Addr().String()
	proxyAddr := proxyListener.Addr().String()

	proxy, err := NewProxy(proxyListener, listenAddr)
	if err != nil {
		return nil, err
	}

	config := config.Config{
		Addr:                 listenAddr,
		CommitLogDir:         "./data",
		CommitLogSegmentSize: 4194304,
	}

	procLogger, err := newLogger(id)
	if err != nil {
		return nil, err
	}

	figg := service.NewFigg(config, procLogger)
	go figg.ServeWithListener(listener)

	return &Node{
		ID:        id,
		Addr:      listenAddr,
		ProxyAddr: proxyAddr,
		proxy:     proxy,
		logger:    logger,
		figg:      figg,
	}, nil
}

func (n *Node) Enable() error {
	if n.proxy != nil {
		n.logger.Debug("node already enabled", zap.String("node-id", n.ID))
		return nil
	}

	proxyListener, err := net.Listen("tcp", n.ProxyAddr)
	if err != nil {
		return err
	}

	proxy, err := NewProxy(proxyListener, n.Addr)
	if err != nil {
		n.logger.Error("failed to enable node", zap.String("node-id", n.ID), zap.Error(err))
		return err
	}
	n.proxy = proxy

	n.logger.Debug("node enabled", zap.String("node-id", n.ID))
	return nil
}

func (n *Node) Disable() error {
	<-time.After(time.Second)  // TODO(AD) tmp

	if err := n.proxy.Close(); err != nil {
		n.logger.Error("failed to disable node", zap.String("node-id", n.ID), zap.Error(err))
	}
	n.proxy = nil

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
	if err := n.proxy.Close(); err != nil {
		return err
	}
	n.figg.Close()
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
