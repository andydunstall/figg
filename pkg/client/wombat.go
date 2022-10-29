package client

import (
	"sync"
	"sync/atomic"
)

type Wombat struct {
	conn       Connection
	messagesCh chan []byte

	wg       sync.WaitGroup
	shutdown int32
}

func NewWombat(addr string, topic string) (*Wombat, error) {
	conn, err := WSConnect(addr, topic, 0)
	if err != nil {
		return nil, err
	}
	wombat := &Wombat{
		conn:       conn,
		messagesCh: make(chan []byte),
		wg:         sync.WaitGroup{},
		shutdown:   0,
	}

	wombat.wg.Add(1)
	go wombat.readLoop()

	return wombat, nil
}

func (w *Wombat) MessagesCh() <-chan []byte {
	return w.messagesCh
}

func (w *Wombat) Shutdown() {
	atomic.StoreInt32(&w.shutdown, 1)

	w.conn.Close()
	// Block until all the listener threads have stopped.
	w.wg.Wait()
}

func (w *Wombat) readLoop() {
	defer w.wg.Done()
	for {
		b, err := w.conn.Recv()
		if err != nil {
			// If we've been shutdown ignore the error.
			if s := atomic.LoadInt32(&w.shutdown); s == 1 {
				return
			}
		}

		msg := b[8:]
		w.messagesCh <- msg
	}
}
