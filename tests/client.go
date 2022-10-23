package tests

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func postMessage(addr string, topic string, message string) error {
	url := fmt.Sprintf("http://%s/v1/%s", addr, topic)
	b := strings.NewReader(message)
	r, err := http.Post(url, "", b)
	if err != nil {
		return err
	}
	if r.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
	return nil
}

type WSClient struct {
	ws *websocket.Conn
}

func WSClientConnect(addr string, topic string, query string) (*WSClient, error) {
	url := fmt.Sprintf("ws://%s/v1/%s/ws?%s", addr, "foo", query)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &WSClient{
		ws: ws,
	}, nil
}

func (c *WSClient) Recv() ([]byte, uint64, error) {
	_, message, err := c.ws.ReadMessage()
	if err != nil {
		return nil, 0, err
	}
	return message, 0, nil
}

func (c *WSClient) Close() {
	c.ws.Close()
}
