package echo

import (
	"errors"
	"net"
	"sync"
)

// Conn wraps an net.Conn and implements its interface.
// See: https://golang.org/pkg/net#Conn
type Conn struct {
	net.Conn
	wg     *sync.WaitGroup
	m      sync.Mutex
	closed bool
}

// Close closes connection.
func (c *Conn) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
		}
	}()

	c.m.Lock()
	if c.closed {
		c.m.Unlock()
		return
	}

	c.wg.Done()
	c.closed = true
	c.m.Unlock()

	return c.Conn.Close()
}
