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

func createWS(addr string, topic string) (*websocket.Conn, error) {
	url := fmt.Sprintf("ws://%s/v1/%s/ws", addr, "foo")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	return ws, err
}
