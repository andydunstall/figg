package server

import (
	"strconv"
	"sync"

	"github.com/andydunstall/figg/service/pkg/conn"
	"github.com/andydunstall/figg/service/pkg/topic"
)

type ClientAttachment struct {
	client *Client
}

func NewClientAttachment(client *Client) topic.Attachment {
	return &ClientAttachment{
		client: client,
	}
}

func (a *ClientAttachment) Send(m topic.Message) {
	a.client.Send(conn.NewPayloadMessage(m.Topic, m.Offset, m.Message))
}

type Client struct {
	conn          conn.Connection
	broker        *topic.Broker
	subscriptions *topic.Subscriptions
	outgoing      [][]byte
	mu            *sync.Mutex
	cv            *sync.Cond
}

func NewClient(conn conn.Connection, broker *topic.Broker) *Client {
	mu := &sync.Mutex{}
	c := &Client{
		conn:     conn,
		broker:   broker,
		outgoing: [][]byte{},
		mu:       mu,
		cv:       sync.NewCond(mu),
	}
	c.subscriptions = topic.NewSubscriptions(broker, NewClientAttachment(c))
	go c.writeLoop()
	return c
}

func (c *Client) Send(m *conn.ProtocolMessage) error {
	b, err := m.Encode()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.outgoing = append(c.outgoing, b)
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

		c.handleIncoming(m)
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
			if err := c.conn.Send(b); err != nil {
				// If we get an error expect the read will fail so the
				// connection will close.
				return
			}
		}
	}
}

func (c *Client) handleIncoming(m *conn.ProtocolMessage) {
	switch m.Type {
	case conn.TypePing:
		c.Send(conn.NewPongMessage(m.Ping.Timestamp))
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
		c.Send(conn.NewAttachedMessage(m.Attach.Topic))
	case conn.TypeDetach:
		// TODO(AD) Unsubscribe
		c.Send(conn.NewDetachedMessage())
	case conn.TypePublish:
		topic, err := c.broker.GetTopic(m.Publish.Topic)
		if err != nil {
			// TODO(AD)
		}
		if err := topic.Publish(m.Publish.Payload); err != nil {
			// TODO(AD)
		}
		c.Send(conn.NewACKMessage(m.Publish.SeqNum))
	}
}

func (c *Client) takeOutgoing() [][]byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	outgoing := c.outgoing
	c.outgoing = [][]byte{}
	return outgoing
}
