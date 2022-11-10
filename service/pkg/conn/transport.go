package conn

import (
	"sync"
	"sync/atomic"
)

type Transport struct {
	conn Connection

	messageCh chan *ProtocolMessage

	wg       sync.WaitGroup
	shutdown int32
}

func NewTransport(conn Connection) *Transport {
	transport := &Transport{
		conn:      conn,
		messageCh: make(chan *ProtocolMessage),
		wg:        sync.WaitGroup{},
		shutdown:  0,
	}

	transport.wg.Add(1)
	go transport.recvLoop()

	return transport
}

func (t *Transport) Send(m *ProtocolMessage) error {
	return t.conn.Send(m)
}

func (t *Transport) MessageCh() <-chan *ProtocolMessage {
	return t.messageCh
}

func (t *Transport) Shutdown() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&t.shutdown, 1)

	// Close the conn, which will stop the read loop.
	if t.conn != nil {
		t.conn.Close()
	}

	// Block until all the listener threads have died.
	t.wg.Wait()
	return nil
}

func (t *Transport) recvLoop() {
	defer t.wg.Done()

	for {
		if s := atomic.LoadInt32(&t.shutdown); s == 1 {
			return
		}

		m, err := t.conn.Recv()
		if err != nil {
			// If we've been shutdown ignore the error and exit.
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				return
			}

			// Return a message of nil to tell the server to close the
			// connection.
			t.messageCh <- nil
			return
		}

		t.messageCh <- m
	}

}
