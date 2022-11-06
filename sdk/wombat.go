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
	subscribers     *Subscribers
	pending         *PendingMessages

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

func (w *Wombat) Publish(topic string, m []byte) {
	// Not worrying about retries yet.
	if err := w.transport.Send(NewPublishMessage(topic, m)); err != nil {
		w.pending.Push(NewPublishMessage(topic, m))
	}
}

func (w *Wombat) Subscribe(topic string, sub MessageSubscriber) {
	w.subscribers.Add(topic, sub)

	// TODO(AD) If already attach just add subscriber.

	// w.subscribers[topic][sub] = nil

	// TODO(AD) Just trying to view logs for incoming messages for now not
	// adding subscriber.
	// TODO(AD) Only if this is the first subscription for the topic.

	w.transport.Send(NewAttachMessage(topic))
}

func (w *Wombat) Shutdown() error {
	if err := w.transport.Shutdown(); err != nil {
		return err
	}
	close(w.doneCh)
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
		subscribers:     NewSubscribers(),
		pending:         NewPendingMessages(),
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
		case m := <-w.transport.MessageCh():
			w.onMessage(m)
		case state := <-w.transport.StateCh():
			w.onConnState(state)

			if w.stateSubscriber != nil {
				w.stateSubscriber.NotifyState(state)
			}
		case <-pingTicker.C:
			w.ping()
		case <-w.doneCh:
			return
		}
	}
}

func (w *Wombat) onMessage(m *ProtocolMessage) {
	w.logger.Debug(
		"on message",
		zap.String("type", TypeToString(m.Type)),
	)

	switch m.Type {
	case TypePayload:
		w.subscribers.OnMessage(m.Payload.Topic, m.Payload.Message)
	}
}

func (w *Wombat) onConnState(s State) {
	w.logger.Debug(
		"on conn state",
		zap.String("type", StateToString(s)),
	)

	switch s {
	case StateConnected:
		w.onConnected()
	}
}

func (w *Wombat) onConnected() {
	for _, m := range w.pending.Get() {
		w.transport.Send(m)
	}

	// Reattach all subscribed topics on connected.
	for _, topic := range w.subscribers.Topics() {
		w.transport.Send(NewAttachMessage(topic))
	}
}

func (w *Wombat) ping() {
	timestamp := time.Now()
	w.logger.Debug("sending ping", zap.Time("timestamp", timestamp))
	m := NewPingMessage(timestamp.UnixMilli())
	// Ignore any errors. If we're not connected we can't send a ping.
	w.transport.Send(m)
}
