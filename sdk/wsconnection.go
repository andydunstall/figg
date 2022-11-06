package figg

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type WSConnection struct {
	ws *websocket.Conn
}

func WSConnect(addr string) (*WSConnection, error) {
	url := fmt.Sprintf("ws://%s/v1/ws", addr)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &WSConnection{
		ws: ws,
	}, nil
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
