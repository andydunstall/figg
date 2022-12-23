package fcm

import (
	"net"
)

type Proxy struct {
	targetAddr string
	listener   net.Listener
	conns      map[*Connection]interface{}
}

func NewProxy(listener net.Listener, targetAddr string) (*Proxy, error) {
	p := &Proxy{
		targetAddr: targetAddr,
		listener:   listener,
		conns:      make(map[*Connection]interface{}),
	}
	go p.acceptLoop()
	return p, nil
}

func (p *Proxy) Close() error {
	if err := p.listener.Close(); err != nil {
		return err
	}

	for conn, _ := range p.conns {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Proxy) acceptLoop() {
	for {
		downstream, err := p.listener.Accept()
		if err != nil {
			return
		}

		upstream, err := net.Dial("tcp", p.targetAddr)
		if err != nil {
			return
		}

		conn := NewConnection(upstream, downstream, p)
		p.conns[conn] = nil
	}
}

func (p *Proxy) removeConn(c *Connection) {
	delete(p.conns, c)
}
