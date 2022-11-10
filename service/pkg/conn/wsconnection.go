package conn

import (
	"github.com/gorilla/websocket"
)

type WSConnection struct {
	ws *websocket.Conn
}

func NewWSConnection(ws *websocket.Conn) Connection {
	return &WSConnection{
		ws: ws,
	}
}

func (t *WSConnection) Send(m *ProtocolMessage) error {
	b, err := m.Encode()
	if err != nil {
		return err
	}
	return t.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (t *WSConnection) Recv() (*ProtocolMessage, error) {
	_, b, err := t.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	m, err := ProtocolMessageFromBytes(b)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (t *WSConnection) Close() error {
	return t.ws.Close()
}
