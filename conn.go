package pool

import (
	"io"
	"sync"
)

// Conn is a wrapper around io.Closer
type Conn struct {
	Client   io.Closer
	mu       sync.RWMutex
	c        *Pool
	unusable bool
}

// Close puts the given connects back to the pool instead of closing it.
func (p *Conn) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.Client != nil {
			return p.Client.Close()
		}
		return nil
	}
	return p.c.put(p.Client)
}

// MarkUnusable () marks the connection not usable any more, to let the pool close it instead of returning it to pool.
func (p *Conn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

// newConn wraps a standard io.Closer to a Conn io.Closer.
func (c *Pool) wrapConn(conn io.Closer) io.Closer {
	p := &Conn{c: c}
	p.Client = conn
	return p
}
