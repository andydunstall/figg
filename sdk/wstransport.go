package wombat

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

func (t *WSTransport) Send(m *ProtocolMessage) error {
	b, err := m.Encode()
	if err != nil {
		return err
	}
	return t.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (t *WSTransport) Recv() (*ProtocolMessage, error) {
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

func (t *WSTransport) Close() error {
	return t.ws.Close()
}
