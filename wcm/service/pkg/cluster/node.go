package cluster

import (
	"fmt"
	"os"
	"sync"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/andydunstall/wombat/service"
	"github.com/andydunstall/wombat/service/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Node struct {
	ID    string
	Addr  string
	proxy *toxiproxy.Proxy

	doneCh chan interface{}
	wg     sync.WaitGroup
}

func NewNode(portAllocator *PortAllocator, toxiproxyClient *toxiproxy.Client) (*Node, error) {
	id := uuid.New().String()

	listenAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())
	proxyAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())

	proxyID := fmt.Sprintf("wombat_%s", id)
	proxy, err := toxiproxyClient.CreateProxy(proxyID, proxyAddr, listenAddr)
	if err != nil {
		return nil, err
	}

	gossipAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())

	config := config.Config{
		Addr:         listenAddr,
		GossipAddr:   gossipAddr,
		GossipPeerID: id,
	}

	logger, err := newLogger(id)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	doneCh := make(chan interface{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		service.Run(config, logger, doneCh)
	}()

	return &Node{
		ID:     id,
		Addr:   proxyAddr,
		proxy:  proxy,
		doneCh: doneCh,
		wg:     wg,
	}, nil
}

func (n *Node) Shutdown() error {
	if err := n.proxy.Delete(); err != nil {
		return err
	}
	close(n.doneCh)
	n.wg.Wait()
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
