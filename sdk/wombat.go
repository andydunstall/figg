package wombat

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type Event int

const (
	EventConnected = Event(1)
)

type Wombat struct {
	addr            string
	transport       Transport
	connectAttempts int

	stateSubscriber StateSubscriber

	wg       sync.WaitGroup
	shutdown int32

	logger *zap.Logger
}

func NewWombat(config *Config) *Wombat {
	wombat := newWombat(config)

	wombat.wg.Add(1)
	go wombat.eventLoop()

	return wombat
}

func (w *Wombat) Shutdown() {
	atomic.StoreInt32(&w.shutdown, 1)

	if w.transport != nil {
		w.transport.Close()
	}
	// Block until all the listener threads have stopped.
	w.wg.Wait()
}

func newWombat(config *Config) *Wombat {
	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	// TODO(AD) check Addr defined

	return &Wombat{
		addr:            config.Addr,
		transport:       nil,
		connectAttempts: 0,
		stateSubscriber: config.StateSubscriber,
		wg:              sync.WaitGroup{},
		shutdown:        0,
		logger:          logger,
	}
}

func (w *Wombat) eventLoop() {
	defer w.wg.Done()

	for {
		if s := atomic.LoadInt32(&w.shutdown); s == 1 {
			return
		}

		if !w.ensureConnected() {
			continue
		}
	}
}

func (w *Wombat) ensureConnected() bool {
	if w.transport != nil {
		return true
	}
	return w.connect()
}

func (w *Wombat) connect() bool {
	backoff := w.getConnectBackoff()

	w.logger.Debug(
		"attempting to connect",
		zap.String("addr", w.addr),
		zap.Duration("backoff", backoff),
	)

	// TODO(AD) check for shutdown
	<-time.After(backoff)

	transport, err := WSTransportConnect(w.addr)
	if err != nil {
		w.logger.Warn("connect failed", zap.Error(err))
		w.connectAttempts++
		return false
	}
	w.connectAttempts = 0
	w.transport = transport

	w.logger.Debug("client connected", zap.String("addr", w.addr))

	if w.stateSubscriber != nil {
		w.stateSubscriber.NotifyState(StateConnected)
	}

	return true
}

func (w *Wombat) getConnectBackoff() time.Duration {
	if w.connectAttempts == 0 {
		return 0
	}

	coefficient := int(math.Pow(float64(2), float64(w.connectAttempts-1)))
	if coefficient > 100 {
		coefficient = 100
	}
	return time.Duration(coefficient) * 100 * time.Millisecond
}
