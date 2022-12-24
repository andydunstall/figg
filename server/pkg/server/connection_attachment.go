package server

import (
	"context"

	"github.com/andydunstall/figg/server/pkg/topic"
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
	c.conn.SendDataMessage(m)
}
