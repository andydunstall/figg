package server

import (
	"context"
	"encoding/binary"
	"strconv"
	"sync"
	"time"

	"github.com/andydunstall/figg/service/pkg/topic"
	"github.com/andydunstall/figg/utils"
)

func RequestContext(requestType utils.MessageType) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "timestamp", time.Now().UnixNano())
	ctx = context.WithValue(ctx, "type", requestType)
	return ctx
}

type ClientAttachment struct {
	client *Client
}

func NewClientAttachment(client *Client) topic.Attachment {
	return &ClientAttachment{
		client: client,
	}
}

func (a *ClientAttachment) Send(ctx context.Context, m topic.Message) {
	// topicPrefix := make([]byte, 2)
	// binary.BigEndian.PutUint16(topicPrefix, uint16(len(m.Topic)))

	// offsetPrefix := make([]byte, 2)
	// binary.BigEndian.PutUint16(offsetPrefix, uint16(len(m.Offset)))

	// messagePrefix := make([]byte, 4)
	// binary.BigEndian.PutUint32(messagePrefix, uint32(len(m.Message)))

	// header := make([]byte, 8)
	// binary.BigEndian.PutUint16(header[:2], uint16(utils.TypePayload))

	// messageLen := uint32(2 + len(m.Topic) + 2 + len(m.Offset) + 4 + len(m.Message))

	buf := []byte{}
	// buf = append(buf, utils.MessageHeader(utils.TypePayload, messageLen)...)
	// buf = append(buf, topicPrefix...)
	// buf = append(buf, []byte(m.Topic)...)
	// buf = append(buf, offsetPrefix...)
	// buf = append(buf, []byte(m.Offset)...)
	// buf = append(buf, messagePrefix...)
	// buf = append(buf, []byte(m.Message)...)

	a.client.Send(ctx, buf)
}

type Outgoing struct {
	Ctx context.Context
	Buf []byte
}

type Client struct {
	conn          utils.Connection
	broker        *topic.Broker
	subscriptions *topic.Subscriptions
	outgoing      []Outgoing
	mu            *sync.Mutex
	cv            *sync.Cond
}

func NewClient(conn utils.Connection, broker *topic.Broker) *Client {
	mu := &sync.Mutex{}
	c := &Client{
		conn:     conn,
		broker:   broker,
		outgoing: []Outgoing{},
		mu:       mu,
		cv:       sync.NewCond(mu),
	}
	c.subscriptions = topic.NewSubscriptions(broker, NewClientAttachment(c))
	go c.writeLoop()
	return c
}

func (c *Client) Send(ctx context.Context, b []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.outgoing = append(c.outgoing, Outgoing{
		Ctx: ctx,
		Buf: b,
	})
	c.cv.Signal()
	return nil
}

func (c *Client) Serve() {
	c.readLoop()
}

func (c *Client) Shutdown() {
	c.subscriptions.Shutdown()
	c.conn.Close()
	// TODO(AD) wait for loops to shut
}

func (c *Client) readLoop() error {
	for {
		messageType, b, err := c.conn.Recv()
		if err != nil {
			// Assume the connection is closed so just return from the
			// read loop (which will cause the client to shutdown).
			return err
		}
		if err = c.handleMessage(messageType, b); err != nil {
			return err
		}
	}
}

func (c *Client) handleMessage(messageType utils.MessageType, b []byte) error {
	switch messageType {
	case utils.TypeAttach:
		return c.handleAttachMessage(b)
	case utils.TypePublish:
		return c.handlePublishMessage(b)
	}
	return nil
}

func (c *Client) handleAttachMessage(b []byte) error {
	ctx := RequestContext(utils.TypeAttach)

	topicLen := binary.BigEndian.Uint16(b[0:2])
	topic := string(b[2 : 2+topicLen])

	offsetLen := binary.BigEndian.Uint16(b[2+topicLen : 2+topicLen+2])
	offsetStr := string(b[2+topicLen+2 : 2+topicLen+2+offsetLen])

	if offsetStr != "" {
		offset, err := strconv.ParseUint(offsetStr, 10, 64)
		if err != nil {
			// If the offset is invalid subscribe without.
			c.subscriptions.AddSubscription(topic)
		} else {
			c.subscriptions.AddSubscriptionFromOffset(topic, offset)
		}
	} else {
		c.subscriptions.AddSubscription(topic)
	}

	topicPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(topicPrefix, uint16(len(topic)))

	messageLen := uint32(2 + len(topic))

	buf := []byte{}
	buf = append(buf, utils.MessageHeader(utils.TypeAttached, messageLen)...)
	buf = append(buf, topicPrefix...)
	buf = append(buf, []byte(topic)...)

	c.Send(ctx, buf)

	return nil
}

func (c *Client) handlePublishMessage(b []byte) error {
	ctx := RequestContext(utils.TypePublish)

	offset := 0

	topicLen := binary.BigEndian.Uint16(b[offset : offset+2])
	offset += 2
	topicName := string(b[offset : offset+int(topicLen)])
	offset += int(topicLen)

	seqNum := binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8

	payloadLen := binary.BigEndian.Uint32(b[offset : offset+4])
	offset += 4
	payload := b[offset : offset+int(payloadLen)]

	topic, err := c.broker.GetTopic(topicName)
	if err != nil {
		// TODO(AD)
	}
	if err := topic.Publish(payload); err != nil {
		// TODO(AD)
	}

	seqNumEnc := make([]byte, 8)
	binary.BigEndian.PutUint64(seqNumEnc, seqNum)

	buf := []byte{}
	buf = append(buf, utils.MessageHeader(utils.TypeACK, 8)...)
	buf = append(buf, seqNumEnc...)

	c.Send(ctx, buf)

	return nil
}

func (c *Client) writeLoop() {
	for {
		c.mu.Lock()
		// Only block if we don't have any outgoing messagese to process
		// (otherwise we can miss signals and deadlock).
		if len(c.outgoing) == 0 {
			c.cv.Wait()
		}
		c.mu.Unlock()

		outgoing := c.takeOutgoing()
		for _, b := range outgoing {
			if err := c.conn.Send(b.Buf); err != nil {
				// If we get an error expect the read will fail so the
				// connection will close.
				return
			}
		}
	}
}

func (c *Client) takeOutgoing() []Outgoing {
	c.mu.Lock()
	defer c.mu.Unlock()

	outgoing := c.outgoing
	c.outgoing = []Outgoing{}
	return outgoing
}
