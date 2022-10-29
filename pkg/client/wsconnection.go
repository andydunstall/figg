package client

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type WSConnection struct {
	ws *websocket.Conn
}

func WSConnect(addr string, topic string, offset uint64) (*WSConnection, error) {
	url := fmt.Sprintf("ws://%s/v1/%s/ws?offset=%d", addr, topic, offset)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &WSConnection{
		ws: ws,
	}, nil
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
