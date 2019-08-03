package pool

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

// Pool implements the Pool interface based on buffered channels.
type Pool struct {
	// storage for our io.Closer connections
	mu    sync.Mutex
	conns chan io.Closer

	// io.Closer generator
	factory Factory
}

// Factory is a function to create new connections.
type Factory func() (io.Closer, error)

// NewPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewPool(initialCap, maxCap int, factory Factory) (Ipool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &Pool{
		conns:   make(chan io.Closer, maxCap),
		factory: factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	return c, nil
}

func (c *Pool) getConnsAndFactory() (chan io.Closer, Factory) {
	c.mu.Lock()
	conns := c.conns
	factory := c.factory
	c.mu.Unlock()
	return conns, factory
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *Pool) Get() (io.Closer, error) {
	conns, factory := c.getConnsAndFactory()
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

		return c.wrapConn(conn), nil
	default:
		conn, err := factory()
		if err != nil {
			return nil, err
		}

		return c.wrapConn(conn), nil
	}
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *Pool) put(conn io.Closer) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conns == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		return nil
	default:
		// pool is full, close passed connection
		return conn.Close()
	}
}

// Close close all resource in the pool
func (c *Pool) Close() error {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

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
func (c *Pool) Len() int {
	conns, _ := c.getConnsAndFactory()
	return len(conns)
}
