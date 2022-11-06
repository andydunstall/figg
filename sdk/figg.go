package figg

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Figg struct {
	transport    *Transport
	pingInterval time.Duration

	stateSubscriber StateSubscriber
	topics          *Topics
	pending         *PendingMessages

	doneCh chan interface{}
	wg     sync.WaitGroup

	logger *zap.Logger
}

func NewFigg(config *Config) (*Figg, error) {
	figg, err := newFigg(config)
	if err != nil {
		return nil, err
	}

	figg.wg.Add(1)
	go figg.eventLoop()

	return figg, nil
}

func (w *Figg) Publish(topic string, m []byte) {
	// Not worrying about retries yet.
	if err := w.transport.Send(NewPublishMessage(topic, m)); err != nil {
		w.pending.Push(NewPublishMessage(topic, m))
	}
}

func (w *Figg) Subscribe(topic string, sub MessageSubscriber) {
	if w.topics.Subscribe(topic, sub) {
		// Note if we arn't connected this won't send an attach, but once
		// we become suspended we attach to all subscribed channels.
		w.transport.Send(NewAttachMessage(topic))
	}
}

func (w *Figg) Unsubscribe(topic string, sub MessageSubscriber) {
	w.topics.Unsubscribe(topic, sub)
}

func (w *Figg) Shutdown() error {
	if err := w.transport.Shutdown(); err != nil {
		return err
	}
	close(w.doneCh)
	w.wg.Wait()
	return nil
}

func newFigg(config *Config) (*Figg, error) {
	if config.Addr == "" {
		return nil, fmt.Errorf("config missing figg address")
	}

	pingInterval := config.PingInterval
	if pingInterval == time.Duration(0) {
		pingInterval = time.Second * 5
	}

	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Figg{
		transport:       NewTransport(config.Addr, logger),
		pingInterval:    pingInterval,
		stateSubscriber: config.StateSubscriber,
		topics:          NewTopics(),
		pending:         NewPendingMessages(),
		doneCh:          make(chan interface{}),
		wg:              sync.WaitGroup{},
		logger:          logger,
	}, nil
}

func (w *Figg) eventLoop() {
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

func (w *Figg) onMessage(m *ProtocolMessage) {
	w.logger.Debug(
		"on message",
		zap.String("type", TypeToString(m.Type)),
	)

	switch m.Type {
	case TypePayload:
		w.topics.OnMessage(m.Payload.Topic, m.Payload.Message)
	}
}

func (w *Figg) onConnState(s State) {
	w.logger.Debug(
		"on conn state",
		zap.String("type", StateToString(s)),
	)

	switch s {
	case StateConnected:
		w.onConnected()
	}
}

func (w *Figg) onConnected() {
	for _, m := range w.pending.Get() {
		w.transport.Send(m)
	}

	// Reattach all subscribed topics on connected.
	for _, topic := range w.topics.Topics() {
		w.transport.Send(NewAttachMessage(topic))
	}
}

func (w *Figg) ping() {
	timestamp := time.Now()
	w.logger.Debug("sending ping", zap.Time("timestamp", timestamp))
	m := NewPingMessage(timestamp.UnixMilli())
	// Ignore any errors. If we're not connected we can't send a ping.
	w.transport.Send(m)
}
