package server

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/andydunstall/figg/service/pkg/conn"
	"github.com/andydunstall/figg/service/pkg/topic"
)

func RequestContext(requestType conn.MessageType) context.Context {
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
	a.client.Send(ctx, conn.NewPayloadMessage(m.Topic, m.Offset, m.Message))
}

type Outgoing struct {
	Ctx context.Context
	Buf []byte
}

type Client struct {
	conn          conn.Connection
	broker        *topic.Broker
	subscriptions *topic.Subscriptions
	outgoing      []Outgoing
	mu            *sync.Mutex
	cv            *sync.Cond
}

func NewClient(conn conn.Connection, broker *topic.Broker) *Client {
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

func (c *Client) Send(ctx context.Context, m *conn.ProtocolMessage) error {
	b, err := m.Encode()
	if err != nil {
		return err
	}

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
		b, err := c.conn.Recv()
		if err != nil {
			// Assume the connection is closed so just return from the
			// read loop (which will cause the client to shutdown).
			return err
		}

		m, err := conn.ProtocolMessageFromBytes(b)
		if err != nil {
			return err
		}

		ctx := RequestContext(m.Type)

		c.handleIncoming(ctx, m)
	}
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

func (c *Client) handleIncoming(ctx context.Context, m *conn.ProtocolMessage) {
	switch m.Type {
	case conn.TypePing:
		c.Send(ctx, conn.NewPongMessage(m.Ping.Timestamp))
	case conn.TypeAttach:
		if m.Attach.Offset != "" {
			offset, err := strconv.ParseUint(m.Attach.Offset, 10, 64)
			if err != nil {
				// If the offset is invalid subscribe without.
				c.subscriptions.AddSubscription(m.Attach.Topic)
			} else {
				c.subscriptions.AddSubscriptionFromOffset(m.Attach.Topic, offset)
			}
		} else {
			c.subscriptions.AddSubscription(m.Attach.Topic)
		}
		c.Send(ctx, conn.NewAttachedMessage(m.Attach.Topic))
	case conn.TypeDetach:
		// TODO(AD) Unsubscribe
		c.Send(ctx, conn.NewDetachedMessage())
	case conn.TypePublish:
		topic, err := c.broker.GetTopic(m.Publish.Topic)
		if err != nil {
			// TODO(AD)
		}
		if err := topic.Publish(m.Publish.Payload); err != nil {
			// TODO(AD)
		}
		message := conn.NewACKMessage(m.Publish.SeqNum)
		c.Send(ctx, message)
	}
}

func (c *Client) takeOutgoing() []Outgoing {
	c.mu.Lock()
	defer c.mu.Unlock()

	outgoing := c.outgoing
	c.outgoing = []Outgoing{}
	return outgoing
}
