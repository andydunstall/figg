package server

import (
	"encoding/binary"

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
	b := make([]byte, 8+len(m))
	binary.BigEndian.PutUint64(b, offset)
	for i := 0; i != len(m); i++ {
		b[i+8] = m[i]
	}
	return c.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (c *WSConn) Recv() ([]byte, error) {
	_, message, err := c.ws.ReadMessage()
	return message, err
}
