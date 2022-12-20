package server

import (
	"context"
	"io"

	"github.com/andydunstall/figg/service/pkg/topic"
	"github.com/andydunstall/figg/utils"
)

const (
	readBufferLen = 1 << 15 // 32 KB
)

type ConnectionAttachment struct {
	conn *Connection
}

func NewConnectionAttachment(conn *Connection) topic.Attachment {
	return &ConnectionAttachment{
		conn: conn,
	}
}

func (c *ConnectionAttachment) Send(ctx context.Context, m topic.Message) {
	c.conn.conn.Write(utils.EncodeDataMessage(m.Topic, m.Offset, m.Message))
}

type NetworkConnection interface {
	io.Reader
	io.Writer
	io.Closer
}

// Connection represents an application level connection to the client.
type Connection struct {
	conn NetworkConnection
	// reader reads bytes from the connection.
	reader *utils.BufferedReader
	// pending contains bytes read from the connection that have not been
	// processed.
	pending []byte

	broker        *topic.Broker
	subscriptions *topic.Subscriptions
}

func NewConnection(conn NetworkConnection, broker *topic.Broker) *Connection {
	c := &Connection{
		conn:   conn,
		reader: utils.NewBufferedReader(conn, readBufferLen),
		broker: broker,
	}
	c.subscriptions = topic.NewSubscriptions(broker, NewConnectionAttachment(c))
	return c
}

// Recv reads from the network connection and handles the request.
func (c *Connection) Recv() error {
	pendingRemaining := c.processPending()

	b, err := c.reader.Read()
	if err != nil {
		return err
	}

	// If there are pending bytes to process, append and process in the next
	// loop.
	if pendingRemaining {
		c.pending = append(c.pending, b...)
		return nil
	}

	messageType, payloadLen, ok := utils.DecodeHeader(b)
	if !ok {
		c.pending = append(c.pending, b...)
		return nil
	}

	if len(b) < utils.HeaderLen+payloadLen {
		c.pending = append(c.pending, b...)
		return nil
	}

	offset := c.onMessage(messageType, utils.HeaderLen, b)
	if offset != len(b) {
		c.pending = append(c.pending, b[offset:]...)
	}

	return nil
}

func (c *Connection) Close() error {
	return nil
}

func (c *Connection) onMessage(messageType utils.MessageType, offset int, b []byte) int {
	switch messageType {
	case utils.TypeAttach:
		flags, offset := utils.DecodeUint16(b, offset)
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)
		topicOffset, offset := utils.DecodeUint64(b, offset)

		if flags&utils.FlagUseOffset > 0 {
			c.onAttachFromOffset(topicName, topicOffset)
		} else {
			c.onAttach(topicName)
		}

		return offset
	case utils.TypePublish:
		topicLen, offset := utils.DecodeUint32(b, offset)
		topicName := string(b[offset : offset+int(topicLen)])
		offset += int(topicLen)
		seqNum, offset := utils.DecodeUint64(b, offset)
		dataLen, offset := utils.DecodeUint32(b, offset)
		data := b[offset : offset+int(dataLen)]
		offset += int(dataLen)

		topic, err := c.broker.GetTopic(topicName)
		if err != nil {
			// TODO(AD)
		}
		if err := topic.Publish(data); err != nil {
			// TODO(AD)
		}

		c.conn.Write(utils.EncodeACKMessage(seqNum))

		return offset
	}

	return 0
}

func (c *Connection) onAttach(name string) {
	c.subscriptions.AddSubscription(name)

	// TODO(AD) include offset
	c.conn.Write(utils.EncodeAttachedMessage(name, 0))
}

func (c *Connection) onAttachFromOffset(name string, offset uint64) {
	c.subscriptions.AddSubscriptionFromOffset(name, offset)

	// TODO(AD) include offset
	c.conn.Write(utils.EncodeAttachedMessage(name, 0))
}

func (c *Connection) processPending() bool {
	if len(c.pending) == 0 {
		return false
	}

	messageType, payloadLen, ok := utils.DecodeHeader(c.pending)
	if !ok {
		return true
	}

	if len(c.pending) < utils.HeaderLen+payloadLen {
		return true
	}

	offset := c.onMessage(messageType, utils.HeaderLen, c.pending)
	c.pending = c.pending[offset:]

	return len(c.pending) == 0
}
