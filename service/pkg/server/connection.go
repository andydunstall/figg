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

// ConnectionAttachment implements a topic attachment to send messages to the
// connection.
type ConnectionAttachment struct {
	conn *Connection
}

func NewConnectionAttachment(conn *Connection) topic.Attachment {
	return &ConnectionAttachment{
		conn: conn,
	}
}

func (c *ConnectionAttachment) Send(ctx context.Context, m topic.Message) {
	// TODO(AD) Can't block, write to background thread.
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
	// reader reads messages from the connection.
	reader *utils.BufferedReader

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
	messageType, payload, err := c.reader.Read()
	if err != nil {
		return err
	}

	c.onMessage(messageType, payload)
	return nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) onMessage(messageType utils.MessageType, b []byte) int {
	offset := 0
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
