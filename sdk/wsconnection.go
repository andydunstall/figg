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

func (c *WSConnection) Send(b []byte) error {
	return c.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (c *WSConnection) Recv() ([]byte, error) {
	_, b, err := c.ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *WSConnection) Close() error {
	return c.ws.Close()
}
