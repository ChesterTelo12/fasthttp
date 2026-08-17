package client

import (
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

type clientConn struct {
	c net.Conn
}

type HostClient struct {
	conns     []*clientConn
	connsLock sync.Mutex
	pool      sync.Pool
}

func (c *HostClient) acquireConn() (*clientConn, error) {
	var cc *clientConn

	if c.conns == nil && c.pool.Len() > 0 {
		cc = c.pool.Get().(*clientConn)
	} else {
		cc = &clientConn{}
	}

	if cc == nil {
		cc = &clientConn{}
	}

	if cc.c == nil {
		// Simulate dialing a new connection
		var d net.Conn
		d, _ = net.Dial("tcp", "127.0.0.1:1234")
		cc.c = d
	}

	// If c.conns is nil, update the pool pointer for next time
	if c.conns == nil {
		c.conns = append(c.conns, cc)
	}
	return cc, nil
}

func (c *HostClient) CloseIdleConnections() {
	c.connsLock.Lock()

	var conns []net.Conn
	if len(c.conns) > 0 {
		conns = make([]net.Conn, len(c.conns))
		for i, cc := range c.conns {
			conns[i] = cc.c
		}
		// Invalidate the main slice so other goroutines know to fetch from pool or iterate local slice
		c.conns = nil
	}
	c.connsLock.Unlock()

	for _, conn := range conns {
		conn.Close()
	}
}

func TestHostClientCloseIdleConnectionsRace(t *testing.T) {
	c := &HostClient{}
	c.connsLock.Lock()
	// Initialize with a wrapper holding a real connection
	dialConn, _ := net.Dial("tcp", "127.0.0.1:1234")
	cc := &clientConn{c: dialConn}
	c.conns = []*clientConn{cc}
	c.connsLock.Unlock()

	// Start the CloseIdleConnections routine
	go func() {
		c.CloseIdleConnections()
	}()

	// Give it a moment to snapshot the connections
	// Then, swap the underlying net.Conn to simulate the 'pool recycle' race
	go func() {
		newDial, _ := net.Dial("tcp", "127.0.0.1:1235") // Slightly different port
		c.connsLock.Lock()
		// Re-attach the new dial to the specific wrapper cc
		// Find cc? c.conns is nil, but we need to simulate the 'readLoop' logic.
		// The race description says 'cc.Close()' finds itself.
		// Let's just assume `cc.c` is the target.
		// We need a 'readLoop' behavior on `cc`.
	}()
}