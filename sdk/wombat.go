package wombat

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/andydunstall/wombat/pkg/conn"
	"github.com/gorilla/websocket"
)

type Wombat struct {
	transport  conn.Transport
	messagesCh chan []byte

	wg       sync.WaitGroup
	shutdown int32
}

func NewWombat(addr string, topic string) (*Wombat, error) {
	url := fmt.Sprintf("ws://%s/v1/%s/ws", addr, topic)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	wombat := &Wombat{
		transport:  conn.NewWSTransport(ws),
		messagesCh: make(chan []byte),
		wg:         sync.WaitGroup{},
		shutdown:   0,
	}

	wombat.wg.Add(1)
	go wombat.readLoop()

	return wombat, nil
}

func (w *Wombat) Publish(b []byte) error {
	return w.transport.Send(&conn.ProtocolMessage{
		Type: conn.TypePublishMessage,
		PublishMessage: &conn.PublishMessage{
			Message: b,
		},
	})
}

func (w *Wombat) MessagesCh() <-chan []byte {
	return w.messagesCh
}

func (w *Wombat) Shutdown() {
	atomic.StoreInt32(&w.shutdown, 1)

	w.transport.Close()
	// Block until all the listener threads have stopped.
	w.wg.Wait()
}

func (w *Wombat) readLoop() {
	defer w.wg.Done()
	for {
		m, err := w.transport.Recv()
		if err != nil {
			// If we've been shutdown ignore the error.
			if s := atomic.LoadInt32(&w.shutdown); s == 1 {
				return
			}
		}
		switch m.Type {
		case conn.TypeTopicMessage:
			w.messagesCh <- m.TopicMessage.Message
		}
	}
}
