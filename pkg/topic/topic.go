package topic

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Topic struct {
	subscribers map[*websocket.Conn]interface{}
	mu          sync.Mutex
}

func NewTopic() *Topic {
	return &Topic{
		subscribers: map[*websocket.Conn]interface{}{},
		mu:          sync.Mutex{},
	}
}

func (t *Topic) Publish(mt int, b []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for sub, _ := range t.subscribers {
		sub.WriteMessage(mt, b)
	}
}

func (t *Topic) Subscribe(conn *websocket.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[conn] = struct{}{}
}

func (t *Topic) Unsubscribe(conn *websocket.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, conn)
}
