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

	// pendingMessages contains protocol messages that must be acknowledged.
	// On a reconnect all pending messages are retried.
	pendingMessages *MessageQueue
	// seqNum is the sequence number of the next protocol message that requires
	// an acknowledgement.
	seqNum uint64

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
	seqNum := w.seqNum
	w.seqNum++

	message := NewPublishMessage(topic, seqNum, m)
	w.pendingMessages.Push(message, seqNum)

	// Ignore errors. If we fail to send we'll retry on reconnect.
	w.transport.Send(message)
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
		pendingMessages: NewMessageQueue(),
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
		zap.Object("message", m),
	)

	switch m.Type {
	case TypeACK:
		w.logger.Debug("on ack", zap.Uint64("seq-num", m.ACK.SeqNum))
		w.pendingMessages.Acknowledge(m.ACK.SeqNum)
	case TypePayload:
		w.logger.Debug("on payload", zap.String("topic", m.Payload.Topic))
		w.topics.OnMessage(m.Payload.Topic, m.Payload.Message, m.Payload.Offset)
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
	// Reattach all subscribed topics.
	for _, topic := range w.topics.Topics() {
		w.logger.Debug("reattaching", zap.String("topic", topic))
		w.transport.Send(NewAttachMessageWithOffset(topic, w.topics.Offset(topic)))
	}

	// Send all unacknowledged messages.
	for _, m := range w.pendingMessages.Messages() {
		w.transport.Send(m)
	}
}

func (w *Figg) ping() {
	timestamp := time.Now()
	w.logger.Debug("sending ping", zap.Time("timestamp", timestamp))
	m := NewPingMessage(timestamp.UnixMilli())
	// Ignore any errors. If we're not connected we can't send a ping.
	w.transport.Send(m)
}
