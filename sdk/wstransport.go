package wombat

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type WSTransport struct {
	ws *websocket.Conn
}

func WSTransportConnect(addr string) (*WSTransport, error) {
	url := fmt.Sprintf("ws://%s/v1/ws", addr)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &WSTransport{
		ws: ws,
	}, nil
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
