package pool

import (
	"errors"
	"fmt"
	"io"
	"time"
)

// Checker check conn healthy
// type Checker interface {
// 	check(io.Closer) bool
// }

// Factory is a function to create new connections.
type Factory func() (time.Duration, io.Closer, func(io.Closer) bool, error)

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

	p := &Pool{
		conns:   make(chan *Conn, maxCap),
		factory: factory,
		maxSize: maxCap,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		timeout, conn, Checker, err := factory()
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		p.conns <- p.wrapConn(conn, timeout, Checker)
	}

	// monitor pool healthy
	// once := &sync.Once{}

	// onceFunc := func() {
	// 	go func() {
	// 		for {
	// 			if p.Len() < initialCap {

	// 				log.Printf("pool size : %d\n", p.Len())
	// 				timeout, conn, checker, err := factory()
	// 				if err != nil {
	// 					conn.Close()
	// 				}
	// 				p.conns <- p.wrapConn(conn, timeout, checker)
	// 				log.Printf("pool size : %d\n", len(p.conns))
	// 			}
	// 			time.Sleep(time.Millisecond * 500)
	// 		}
	// 	}()
	// }
	// once.Do(onceFunc)
	return p, nil
}

// TODO: pool connection healthcheck
//func poolCapacityCheck(p IPool) {}
