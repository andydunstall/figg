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

	outgoingMessages chan *ProtocolMessage
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
	w.outgoingMessages <- message
}

func (w *Figg) Subscribe(topic string, cb MessageHandler) *MessageSubscriber {
	sub, activated := w.topics.Subscribe(topic, cb)
	if activated {
		// Note if we arn't connected this won't send an attach, but once
		// we become suspended we attach to all subscribed channels.
		w.outgoingMessages <- NewAttachMessage(topic)
	}
	return sub
}

func (w *Figg) Unsubscribe(topic string, sub *MessageSubscriber) {
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

	figg := &Figg{
		pingInterval:     pingInterval,
		stateSubscriber:  config.StateSubscriber,
		topics:           NewTopics(),
		outgoingMessages: make(chan *ProtocolMessage),
		pendingMessages:  NewMessageQueue(),
		doneCh:           make(chan interface{}),
		wg:               sync.WaitGroup{},
		logger:           logger,
	}
	figg.transport = NewTransport(config.Addr, logger, figg.onMessage, figg.onState)
	return figg, nil
}

func (w *Figg) eventLoop() {
	defer w.wg.Done()

	pingTicker := time.NewTicker(w.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case m := <-w.outgoingMessages:
			w.transport.Send(m)
		case <-pingTicker.C:
			w.ping()
		case <-w.doneCh:
			return
		}
	}
}

func (w *Figg) onState(state State) {
	w.onConnState(state)

	if w.stateSubscriber != nil {
		w.stateSubscriber.NotifyState(state)
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
		w.outgoingMessages <- NewAttachMessageWithOffset(topic, w.topics.Offset(topic))
	}

	// Send all unacknowledged messages.
	for _, m := range w.pendingMessages.Messages() {
		w.outgoingMessages <- m
	}
}

func (w *Figg) ping() {
	timestamp := time.Now()
	w.logger.Debug("sending ping", zap.Time("timestamp", timestamp))
	m := NewPingMessage(timestamp.UnixMilli())
	// Ignore any errors. If we're not connected we can't send a ping.
	w.outgoingMessages <- m
}
