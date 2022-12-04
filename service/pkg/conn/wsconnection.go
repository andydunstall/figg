package conn

import (
	"github.com/fasthttp/websocket"
)

type WSConnection struct {
	ws *websocket.Conn
}

func NewWSConnection(ws *websocket.Conn) Connection {
	return &WSConnection{
		ws: ws,
	}
}

func (t *WSConnection) Send(b []byte) error {
	return t.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (t *WSConnection) Recv() ([]byte, error) {
	_, b, err := t.ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (t *WSConnection) Close() error {
	return t.ws.Close()
}
