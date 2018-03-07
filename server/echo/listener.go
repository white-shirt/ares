package echo

import (
	"net"
	"sync"
	"time"
)

// Listener wraps an net.Listener and implements its interface.
// See: https://golang.org/pkg/net/#Listener
type Listener struct {
	net.Listener
	wg *sync.WaitGroup
}

func wrapListener(l net.Listener) *Listener {
	return &Listener{
		Listener: l,
		wg:       &sync.WaitGroup{},
	}
}

// Accept implements the net.Listener interface to allow an net listener to Accept
// waits for and returns the next connection to the listener.
// See: https://golang.org/pkg/net/#Listener
func (l *Listener) Accept() (c net.Conn, err error) {
	tc, err := l.Listener.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return nil, err
	}

	if err = tc.SetKeepAlive(true); err != nil {
		return nil, err
	}

	if err = tc.SetKeepAlivePeriod(20 * time.Second); err != nil {
		return nil, err
	}

	l.wg.Add(1)

	return &Conn{Conn: tc, wg: l.wg}, nil
}
