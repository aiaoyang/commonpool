package pool

import (
	"fmt"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

var (
	f = func() (time.Duration, io.Closer, error) {
		c, err := net.Dial("tcp", "10.0.3.34:80")
		return time.Millisecond, c, err
	}
)

func Test_Conn_Timeout(t *testing.T) {
	p, err := NewPool(5, 5, f)
	if err != nil {
		log.Fatal(err)
	}
	c, err := p.Get()
	time.Sleep(time.Millisecond * 10)
	c.Close()
	fmt.Println(p.Len())
	time.Sleep(time.Millisecond * 10)
	c.Close()
	fmt.Println(p.Len())
	p.Get()
	fmt.Println(p.Len())

}

func Test_conn(t *testing.T) {
	p, err := NewPool(5, 5, f)
	if err != nil {
		log.Fatal(err)
	}
	c, err := p.Get()
	if err != nil {
		log.Fatal(err)
	}
	b := make([]byte, 100)
	c.Client.(net.Conn).Write([]byte("test"))
	c.Client.(net.Conn).Read(b)
	fmt.Printf("%s\n", b)
	c.Close()
	p.Close()
}
func BenchmarkConn(b *testing.B) {
	p, err := NewPool(1, 2, f)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		c, err := p.Get()
		if err != nil {
			log.Fatal(err)
		}
		if i%2 == 0 {
			c.timeout = 0
			fmt.Printf("timeout detected\n")
			time.Sleep(time.Millisecond * 10)
		}
		fmt.Printf("pool's len is : %d\n", p.Len())
		fmt.Printf("conn's address is : %v\n", &c)
		c.Client.(net.Conn).Write([]byte("test"))

		//b := make([]byte, 100)
		//c.Client.(net.Conn).Write([]byte("test"))
		//c.Client.(net.Conn).Read(b)
		//fmt.Printf("%s\n", b)
		c.Close()
		fmt.Printf("pool's len is : %d\n", p.Len())

	}
	//p.Close()
}
