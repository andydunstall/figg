package figg

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	// addr is the address of the figg service.
	addr string
	// conn is the connection to a figg server or nil if not connected.
	conn Connection
	// connectAttempts is the number of attempts to connect to figg without
	// being able to connect.
	connectAttempts int

	messageCb func(m *ProtocolMessage)
	stateCb   func(s State)

	outgoing [][]byte
	mu       *sync.Mutex
	cv       *sync.Cond

	wg       sync.WaitGroup
	shutdown int32

	logger *zap.Logger
}

func NewClient(addr string, logger *zap.Logger, messageCb func(m *ProtocolMessage), stateCb func(s State)) *Client {
	mu := &sync.Mutex{}
	client := &Client{
		addr:            addr,
		conn:            nil,
		connectAttempts: 0,
		messageCb:       messageCb,
		stateCb:         stateCb,
		outgoing:        [][]byte{},
		mu:              mu,
		cv:              sync.NewCond(mu),
		wg:              sync.WaitGroup{},
		shutdown:        0,
		logger:          logger,
	}

	client.wg.Add(1)
	go client.readLoop()
	go client.writeLoop()

	return client
}

func (c *Client) Send(m *ProtocolMessage) error {
	if c.conn == nil {
		return fmt.Errorf("client not connected")
	}

	c.logger.Debug(
		"send message",
		zap.Object("message", m),
	)

	b, err := m.Encode()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.outgoing = append(c.outgoing, b)
	c.cv.Signal()
	return nil
}

func (c *Client) Shutdown() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&c.shutdown, 1)

	// Close the conn, which will stop the read loop.
	if c.conn != nil {
		c.conn.Close()
	}

	// Block until all the listener threads have died.
	c.wg.Wait()
	return nil
}

func (c *Client) readLoop() {
	defer c.wg.Done()

	for {
		if s := atomic.LoadInt32(&c.shutdown); s == 1 {
			return
		}

		if c.conn == nil {
			if err := c.connect(); err != nil {
				continue
			}
			c.stateCb(StateConnected)
		}

		m, err := c.recv()
		if err != nil {
			// If we've been shutdown ignore the error and exit.
			if s := atomic.LoadInt32(&c.shutdown); s == 1 {
				return
			}

			c.logger.Debug("read failed", zap.Error(err))

			c.stateCb(StateDisconnected)
			c.conn = nil

			continue
		}

		c.messageCb(m)
	}
}

func (c *Client) writeLoop() {
	defer c.wg.Done()

	for {
		c.mu.Lock()
		// Only block if we don't have any outgoing messagese to process
		// (otherwise we can miss signals and deadlock).
		if len(c.outgoing) == 0 {
			c.cv.Wait()
		}
		c.mu.Unlock()

		outgoing := c.takeOutgoing()
		for _, b := range outgoing {
			if err := c.conn.Send(b); err != nil {
				// If we get an error expect the read will fail so the
				// connection will close.
				return
			}
		}
	}
}

func (c *Client) connect() error {
	backoff := c.getBackoffTimeout(c.connectAttempts)

	c.logger.Debug(
		"connecting",
		zap.String("addr", c.addr),
		zap.Duration("backoff", backoff),
	)

	<-time.After(backoff)

	conn, err := WSConnect(c.addr)
	if err != nil {
		c.connectAttempts += 1

		c.logger.Debug("connection failed", zap.Error(err))
		return err
	}

	c.conn = conn
	c.connectAttempts = 0

	c.logger.Debug("connection ok")
	return nil
}

func (c *Client) recv() (*ProtocolMessage, error) {
	b, err := c.conn.Recv()
	if err != nil {
		return nil, err
	}

	return ProtocolMessageFromBytes(b)
}

func (c *Client) getBackoffTimeout(n int) time.Duration {
	if n == 0 {
		return 0
	}

	coefficient := int(math.Pow(float64(2), float64(n-1)))
	if coefficient > 100 {
		coefficient = 100
	}
	return time.Duration(coefficient) * 100 * time.Millisecond
}

func (c *Client) takeOutgoing() [][]byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	outgoing := c.outgoing
	c.outgoing = [][]byte{}
	return outgoing
}
