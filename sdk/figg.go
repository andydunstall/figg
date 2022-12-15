package figg

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/andydunstall/figg/utils"
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

	pendingAttaches *AttachQueue

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

// Subscribe to the given topic.
func (w *Figg) Subscribe(ctx context.Context, topic string, cb MessageHandler) (*MessageSubscriber, error) {
	sub, activated := w.topics.Subscribe(topic, cb)
	if activated {
		// Attach to the topic.
		ch := make(chan interface{}, 1)
		w.attach(topic, func() {
			ch <- struct{}{}
		})
		select {
		case <-ch:
			return sub, nil
		case <-ctx.Done():
			return sub, ctx.Err()
		}
	}

	// We're already attached.
	return sub, nil
}

// SubscribeFrom to the given topic from the given offset.
//
// Note this does not work when there are existing subscribers as can't update
// the offset for those subscribers, used for testing only.
func (w *Figg) SubscribeFromOffset(ctx context.Context, topic string, offset string, cb MessageHandler) (*MessageSubscriber, error) {
	sub, activated := w.topics.Subscribe(topic, cb)
	if activated {
		// Attach to the topic.
		ch := make(chan interface{}, 1)
		w.attachFromOffset(topic, offset, func() {
			ch <- struct{}{}
		})
		select {
		case <-ch:
			return sub, nil
		case <-ctx.Done():
			return sub, ctx.Err()
		}
	}
	return sub, nil
}

// Unsubscribe from the given topic.
func (w *Figg) Unsubscribe(topic string, sub *MessageSubscriber) {
	if w.topics.Unsubscribe(topic, sub) {
		// Ignore errors. If we arn't connected so the send fails, when we
		// reconnect this topic won't be reattached.
		// TODO w.client.Send(NewDetachMessage(topic))
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
		pendingAttaches: NewAttachQueue(),
		doneCh:          make(chan interface{}),
		wg:              sync.WaitGroup{},
		logger:          logger,
	}
	figg.client = NewClient(config.Addr, logger, figg.onMessage, figg.onState)
	return figg, nil
}

func (w *Figg) publish(topic string, m []byte, cb func(err error)) {
	// Publish messages must be acknowledged so add a sequence number and queue.
	seqNum := w.seqNum
	w.seqNum++

	buf := utils.PublishMessage(topic, seqNum, m)
	w.pendingMessages.Push(buf, seqNum, cb)

	w.logger.Debug(
		"sending payload message",
		zap.String("topic", topic),
		zap.Uint64("seq-num", seqNum),
		zap.Int("payload-len", len(m)),
	)

	// Ignore errors. If we fail to send we'll retry.
	w.client.SendBytes(buf)
}

func (w *Figg) attach(topic string, cb func()) {
	w.attachFromOffset(topic, "", cb)
}

func (w *Figg) attachFromOffset(topic string, offset string, cb func()) {
	if cb != nil {
		w.pendingAttaches.Push(topic, cb)
	}

	w.logger.Debug(
		"sending attach message",
		zap.String("topic", topic),
		zap.String("offset", offset),
	)

	buf := utils.AttachMessage(topic, offset)

	// Ignore errors. If we arn't connected so the send fails, when we
	// reconnect all subscribed topics are reattached.
	w.client.SendBytes(buf)
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

func (w *Figg) onMessage(messageType utils.MessageType, b []byte) {
	switch messageType {
	case utils.TypePayload:
		w.onPayloadMessage(b)
	case utils.TypeACK:
		w.onACKMessage(b)
	case utils.TypeAttached:
		w.onAttachedMessage(b)
	}
}

func (w *Figg) onPayloadMessage(b []byte) {
	offset := 0

	topicLen := binary.BigEndian.Uint16(b[offset : offset+2])
	offset += 2
	topicName := string(b[offset : offset+int(topicLen)])
	offset += int(topicLen)

	offsetLen := binary.BigEndian.Uint16(b[offset : offset+2])
	offset += 2
	messageOffset := string(b[offset : offset+int(offsetLen)])
	offset += int(offsetLen)

	payloadLen := binary.BigEndian.Uint32(b[offset : offset+4])
	offset += 4
	payload := b[offset : offset+int(payloadLen)]

	w.logger.Debug(
		"on payload message",
		zap.String("topic", topicName),
		zap.String("offset", messageOffset),
		zap.Int("payload-len", len(payload)),
	)

	w.topics.OnMessage(topicName, payload, messageOffset)
}

func (w *Figg) onACKMessage(b []byte) {
	offset := 0

	seqNum := binary.BigEndian.Uint64(b[offset : offset+8])

	w.logger.Debug(
		"on ack message",
		zap.Uint64("seq-num", seqNum),
	)

	w.pendingMessages.Acknowledge(seqNum)
}

func (w *Figg) onAttachedMessage(b []byte) {
	offset := 0

	topicLen := binary.BigEndian.Uint16(b[offset : offset+2])
	offset += 2
	topicName := string(b[offset : offset+int(topicLen)])
	offset += int(topicLen)

	w.logger.Debug(
		"on attached message",
		zap.String("topic", topicName),
	)

	w.pendingAttaches.Attached(topicName)
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
		offset := w.topics.Offset(topic)
		w.logger.Debug(
			"reattaching",
			zap.String("topic", topic),
			zap.String("offset", offset),
		)
		// Ignore errors. If we arn't connected so the send fails, when we
		// reconnect all subscribed topics are reattached.
		w.attachFromOffset(topic, offset, nil)
	}

	// Send all unacknowledged messages.
	for _, m := range w.pendingMessages.Messages() {
		// Ignore errors. If we fail to send we'll retry.
		w.client.SendBytes(m)
	}
}

func (w *Figg) ping() {
	timestamp := time.Now()
	w.logger.Debug("sending ping", zap.Time("timestamp", timestamp))
	m := utils.PingMessage(uint64(timestamp.UnixMilli()))
	// Ignore any errors. If we can't send a ping we'll reconnect anyway.
	w.client.SendBytes(m)
}
