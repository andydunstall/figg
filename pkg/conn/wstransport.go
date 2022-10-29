package conn

import (
	"github.com/gorilla/websocket"
)

type WSTransport struct {
	ws *websocket.Conn
}

func NewWSTransport(ws *websocket.Conn) Transport {
	return &WSTransport{
		ws: ws,
	}
}

func (t *WSTransport) Send(b []byte) error {
	return t.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (t *WSTransport) Recv() ([]byte, error) {
	_, b, err := t.ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (t *WSTransport) Close() error {
	return t.ws.Close()
}
