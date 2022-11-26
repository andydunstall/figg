package figg

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type Transport struct {
	// addr is the address of the figg service.
	addr string
	// conn is the connection to a figg server or nil if not connected.
	conn Connection
	// connectAttempts is the number of attempts to connect to figg without
	// being able to connect.
	connectAttempts int

	messageCb func(m *ProtocolMessage)
	stateCb   func(s State)

	wg       sync.WaitGroup
	shutdown int32

	logger *zap.Logger
}

func NewTransport(addr string, logger *zap.Logger, messageCb func(m *ProtocolMessage), stateCb func(s State)) *Transport {
	transport := &Transport{
		addr:            addr,
		conn:            nil,
		connectAttempts: 0,
		messageCb:       messageCb,
		stateCb:         stateCb,
		wg:              sync.WaitGroup{},
		shutdown:        0,
		logger:          logger,
	}

	transport.wg.Add(1)
	go transport.recvLoop()

	return transport
}

func (t *Transport) Send(m *ProtocolMessage) error {
	if t.conn == nil {
		return fmt.Errorf("transport not connected")
	}

	t.logger.Debug(
		"send message",
		zap.Object("message", m),
	)

	b, err := m.Encode()
	if err != nil {
		return err
	}

	return t.conn.Send(b)
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

		if t.conn == nil {
			if err := t.connect(); err != nil {
				continue
			}
			t.stateCb(StateConnected)
		}

		m, err := t.recv()
		if err != nil {
			// If we've been shutdown ignore the error and exit.
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				return
			}

			t.logger.Debug("read failed", zap.Error(err))

			t.stateCb(StateDisconnected)
			t.conn = nil

			continue
		}

		t.messageCb(m)
	}
}

func (t *Transport) connect() error {
	backoff := t.getBackoffTimeout(t.connectAttempts)

	t.logger.Debug(
		"connecting",
		zap.String("addr", t.addr),
		zap.Duration("backoff", backoff),
	)

	<-time.After(backoff)

	conn, err := WSConnect(t.addr)
	if err != nil {
		t.connectAttempts += 1

		t.logger.Debug("connection failed", zap.Error(err))
		return err
	}

	t.conn = conn
	t.connectAttempts = 0

	t.logger.Debug("connection ok")
	return nil
}

func (t *Transport) recv() (*ProtocolMessage, error) {
	b, err := t.conn.Recv()
	if err != nil {
		return nil, err
	}

	return ProtocolMessageFromBytes(b)
}

func (t *Transport) getBackoffTimeout(n int) time.Duration {
	if n == 0 {
		return 0
	}

	coefficient := int(math.Pow(float64(2), float64(n-1)))
	if coefficient > 100 {
		coefficient = 100
	}
	return time.Duration(coefficient) * 100 * time.Millisecond
}
