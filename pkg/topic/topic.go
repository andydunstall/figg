package topic

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Topic struct {
	subscribers map[string]*websocket.Conn
	mu          sync.Mutex
}

func NewTopic() *Topic {
	return &Topic{
		subscribers: map[string]*websocket.Conn{},
		mu:          sync.Mutex{},
	}
}

func (t *Topic) Publish(mt int, b []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, sub := range t.subscribers {
		sub.WriteMessage(mt, b)
	}
}

func (t *Topic) Subscribe(addr string, conn *websocket.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers[addr] = conn
}

func (t *Topic) Unsubscribe(addr string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscribers, addr)
}
