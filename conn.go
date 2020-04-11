package pool

import (
	"io"
	"sync"
	"time"
)

// Checker check conn healthy
// // var Checker func(io.Closer) bool
// type Checker interface {
// 	check(io.Closer) bool
// }

// Conn is a wrapper around io.Closer
type Conn struct {
	Client     io.Closer
	Checker    func(io.Closer) bool
	mu         sync.RWMutex
	p          *Pool
	activeTime time.Time
	timeout    time.Duration
}

// CheckOK check conn health
func (c *Conn) CheckOK() bool {
	return c.Checker(c.Client)
}

// Close puts the given connects back to the pool instead of closing it.
func (c *Conn) Close() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Client == nil {
		return nil
	}
	if !c.CheckOK() {
		c.Client.Close()
		return nil
	}
	if c.p.Len() == c.p.maxSize {
		c.Client.Close()
		//log.Printf("pool is full,closing connection: %v \n", &c)
		return nil
	}
	// if timeout detected then close it
	if c.activeTime.Add(c.timeout).Before(time.Now()) || c.activeTime.Add(c.timeout).Equal(time.Now()) {
		if c.Client != nil {
			return c.Client.Close()
		}
		return nil
	}
	return c.p.put(c)
}

// MarkUnusable () marks the connection not usable any more, to let the pool close it instead of returning it to pool.
func (c *Conn) MarkUnusable() {
	c.mu.Lock()
	c.timeout = 0
	c.mu.Unlock()
}

// newConn wraps a standard io.Closer to a Conn io.Closer.
