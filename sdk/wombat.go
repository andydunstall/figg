package wombat

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Wombat struct {
	transport    *Transport
	pingInterval time.Duration

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
	if config.Addr == "" {
		return nil, fmt.Errorf("config missing wombat address")
	}

	pingInterval := config.PingInterval
	if pingInterval == time.Duration(0) {
		pingInterval = time.Second * 5
	}

	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Wombat{
		transport:       NewTransport(config.Addr, logger),
		pingInterval:    pingInterval,
		stateSubscriber: config.StateSubscriber,
		doneCh:          make(chan interface{}),
		wg:              sync.WaitGroup{},
		logger:          logger,
	}, nil
}

func (w *Wombat) eventLoop() {
	defer w.wg.Done()

	pingTicker := time.NewTicker(w.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case state := <-w.transport.StateCh():
			if w.stateSubscriber != nil {
				w.stateSubscriber.NotifyState(state)
			}
		case <-pingTicker.C:
			w.logger.Debug("ping")
		case <-w.doneCh:
			return
		}
	}
}
