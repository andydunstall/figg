package server

import (
	"github.com/andydunstall/wombat/pkg/topic"
	"github.com/gorilla/websocket"
)

type WSConn struct {
	ws *websocket.Conn
}

func NewWSConn(ws *websocket.Conn) topic.Conn {
	return &WSConn{
		ws: ws,
	}
}

func (c *WSConn) Send(offset uint64, m []byte) error {
	// TODO(AD) prepend 8 byte offset
	return c.ws.WriteMessage(websocket.BinaryMessage, m)
}

func (c *WSConn) Recv() ([]byte, error) {
	_, message, err := c.ws.ReadMessage()
	return message, err
}
