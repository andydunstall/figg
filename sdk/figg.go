package figg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Figg struct {
	client       *Client
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
	go figg.pingLoop()

	return figg, nil
}

// Publishes publishes the message on the given topic.
//
// This doesn't not wait for the publish to be acknowledged. Similar to TCP,
// Figg guarantees messages are published from the client in order and will
// reconnect/retry if the publish fails, though delivery cannot be guaranteed
// (if the client cannot reconnect).
func (w *Figg) Publish(ctx context.Context, topic string, m []byte) error {
	ch := make(chan error, 1)
	w.publish(topic, m, func(err error) {
		ch <- err
	})
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *Figg) PublishNoACK(topic string, m []byte) {
	w.publish(topic, m, nil)
}

func (w *Figg) publish(topic string, m []byte, cb func(err error)) {
	// Publish messages must be acknowledged so add a sequence number and queue.
	seqNum := w.seqNum
	w.seqNum++
	message := NewPublishMessage(topic, seqNum, m)
	w.pendingMessages.Push(message, seqNum, cb)

	// Ignore errors. If we fail to send we'll retry.
	w.client.Send(message)
}

// Subscribe to the given topic.
func (w *Figg) Subscribe(topic string, cb MessageHandler) *MessageSubscriber {
	sub, activated := w.topics.Subscribe(topic, cb)
	if activated {
		// Ignore errors. If we arn't connected so the send fails, when we
		// reconnect all subscribed topics are reattached.
		w.client.Send(NewAttachMessage(topic))
	}
	return sub
}

// SubscribeFrom to the given topic from the given offset.
//
// Note this does not work when there are existing subscribers as can't update
// the offset for those subscribers, used for testing only.
func (w *Figg) SubscribeFromOffset(topic string, offset string, cb MessageHandler) *MessageSubscriber {
	sub, activated := w.topics.Subscribe(topic, cb)
	if activated {
		// Ignore errors. If we arn't connected so the send fails, when we
		// reconnect all subscribed topics are reattached.
		w.client.Send(NewAttachMessageWithOffset(topic, offset))
	}
	return sub
}

// Unsubscribe from the given topic.
func (w *Figg) Unsubscribe(topic string, sub *MessageSubscriber) {
	if w.topics.Unsubscribe(topic, sub) {
		// Ignore errors. If we arn't connected so the send fails, when we
		// reconnect this topic won't be reattached.
		w.client.Send(NewDetachMessage(topic))
	}
}

func (w *Figg) Shutdown() error {
	if err := w.client.Shutdown(); err != nil {
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
		pingInterval:    pingInterval,
		stateSubscriber: config.StateSubscriber,
		topics:          NewTopics(),
		pendingMessages: NewMessageQueue(),
		doneCh:          make(chan interface{}),
		wg:              sync.WaitGroup{},
		logger:          logger,
	}
	figg.client = NewClient(config.Addr, logger, figg.onMessage, figg.onState)
	return figg, nil
}

func (w *Figg) pingLoop() {
	defer w.wg.Done()

	pingTicker := time.NewTicker(w.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
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
		// Ignore errors. If we arn't connected so the send fails, when we
		// reconnect all subscribed topics are reattached.
		w.client.Send(NewAttachMessageWithOffset(topic, w.topics.Offset(topic)))
	}

	// Send all unacknowledged messages.
	for _, m := range w.pendingMessages.Messages() {
		// Ignore errors. If we fail to send we'll retry.
		w.client.Send(m)
	}
}

func (w *Figg) ping() {
	timestamp := time.Now()
	w.logger.Debug("sending ping", zap.Time("timestamp", timestamp))
	m := NewPingMessage(timestamp.UnixMilli())
	// Ignore any errors. If we can't send a ping we'll reconnect anyway.
	w.client.Send(m)
}
