package pool

import (
	"errors"
	"fmt"
	"io"
	"time"
)

// Factory is a function to create new connections.
type Factory func() (time.Duration, io.Closer, error)

// NewPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewPool(initialCap, maxCap int, factory Factory) (IPool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &Pool{
		conns:   make(chan *Conn, maxCap),
		factory: factory,
		maxSize: maxCap,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		timeout, conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- c.wrapConn(conn, timeout)
	}

	return c, nil
}

// TODO: pool connection healthcheck
//func poolCapacityCheck(p IPool) {}
