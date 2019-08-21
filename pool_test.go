package pool

import (
	"io"
	"net"
	"testing"
	"time"
)

var (
	factory = func() (time.Duration, io.Closer, error) {
		c, err := net.Dial("tcp", "127.0.0.1:10000")
		return time.Duration(10), c, err
	}
)

func Benchmark_GetConn(b *testing.B) {
	l, err := net.Listen("tcp", "127.0.0.1:10000")
	if err != nil {
		b.Errorf("tcp listener err: %s\n", err.Error())
	}
	defer l.Close()
	P, err := NewPool(10, 100, factory)
	if err != nil {
		b.Errorf("create pool err: %s\n", err.Error())
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c, _ := P.Get()
		c.Close()
	}
	b.StopTimer()

}
