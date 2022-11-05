package wombat

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Wombat struct {
	transport *Transport

	stateSubscriber StateSubscriber

	doneCh chan interface{}
	wg     sync.WaitGroup

	logger *zap.Logger
}

func NewWombat(config *Config) (*Wombat, error) {
	wombat, err := newWombat(config)
	if err != nil {
		return nil, err
	}

	wombat.wg.Add(1)
	go wombat.eventLoop()

	return wombat, nil
}

func (w *Wombat) Shutdown() error {
	close(w.doneCh)
	if err := w.transport.Shutdown(); err != nil {
		return err
	}
	w.wg.Wait()
	return nil
}

func newWombat(config *Config) (*Wombat, error) {
	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.Addr == "" {
		return nil, fmt.Errorf("config missing wombat address")
	}

	return &Wombat{
		transport:       NewTransport(config.Addr, logger),
		stateSubscriber: config.StateSubscriber,
		doneCh:          make(chan interface{}),
		wg:              sync.WaitGroup{},
		logger:          logger,
	}, nil
}

func (w *Wombat) eventLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.doneCh:
			return
		case state := <-w.transport.StateCh():
			if w.stateSubscriber != nil {
				w.stateSubscriber.NotifyState(state)
			}
		}
	}
}
