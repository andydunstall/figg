package server

import (
	"github.com/andydunstall/wombat/pkg/topic"
	"github.com/gorilla/websocket"
)

type WSSubscriber struct {
	ws *websocket.Conn
}

func NewWSSubscriber(ws *websocket.Conn) topic.Subscriber {
	return &WSSubscriber{
		ws: ws,
	}
}

func (s *WSSubscriber) Notify(b []byte) {
	// For now this just blocks. In the future to avoid slow updates with high
	// fan out add a background thread to handle writes and just add the message
	// to a outgoing buffer.
	//
	// Note as this blocks we know we arn't doing multiple writes on the same
	// websocket from multiple goroutines.
	s.ws.WriteMessage(websocket.BinaryMessage, b)
}
