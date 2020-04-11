// Package pool implements a pool of net.Conn interfaces to manage and reuse them.
package pool

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// Ipool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
type IPool interface {
	// Get returns a new connection from the pool. Closing the connections puts
	// it back to the Pool. Closing it when the pool is destroyed or full will
	// be counted as an error.
	Get() (*Conn, error)

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close() error

	// Len returns the current number of connections of the pool.
	Len() int
}

// Pool implements the Pool interface based on buffered channels.
type Pool struct {
	// storage for our io.Closer connections
	mu      sync.Mutex
	conns   chan *Conn
	maxSize int

	// io.Closer generator
	factory Factory
}

func (p *Pool) wrapConn(conn io.Closer, timeout time.Duration, checker func(io.Closer) bool) *Conn {
	c := &Conn{
		p:          p,
		Checker:    checker,
		activeTime: time.Now(),
		timeout:    timeout,
	}
	c.Client = conn
	return c
}
func (p *Pool) getConnsAndFactory() (chan *Conn, Factory) {
	p.mu.Lock()
	conns := p.conns
	factory := p.factory
	p.mu.Unlock()
	return conns, factory
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (p *Pool) Get() (*Conn, error) {
	conns, factory := p.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}
		// if conn is timeout, close it and regenerate one
		if conn.activeTime.Add(conn.timeout).Before(time.Now()) {
			//log.Printf("conn is timeout : %v, now closing it\n", &conn)
			conn.Client.Close()
			//log.Printf("conn : %v is closed, now generating new conn\n", &conn)
			timeout, c, checker, err := factory()
			if err != nil {
				return nil, err
			}
			return p.wrapConn(c, timeout, checker), nil
		}
		//log.Printf("return conn : %v\n", &conn)

		return conn, nil
	default:
		timeout, c, checker, err := factory()
		if err != nil {
			return nil, err
		}

		return p.wrapConn(c, timeout, checker), nil
	}
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (p *Pool) put(conn *Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conns == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case p.conns <- conn:
		return nil
	default:
		// pool is full, close passed connection
		return conn.Close()
	}
}

// Close close all resource in the pool
func (p *Pool) Close() error {
	p.mu.Lock()
	conns := p.conns
	p.conns = nil
	p.factory = nil
	p.mu.Unlock()

	if conns == nil {
		return fmt.Errorf("conns is nil")
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
	return nil
}

// Len return pool's length
func (p *Pool) Len() int {
	conns, _ := p.getConnsAndFactory()
	return len(conns)
}
