package listenerutils

import (
	"errors"
	"net"
	"time"
)

//CustomListener provides an implementation for a custom net.Listener
type CustomListener interface {
	net.Listener
	CanClose(bool)
	GetListener() CustomListener
	StopAccept()
}

type customListener struct {
	net.Listener
	canClose bool
	accept   chan *acceptValues
	stop     chan *bool
	timeout  time.Duration
}

type acceptValues struct {
	conn net.Conn
	err  error
}

func (c *customListener) Close() error {
	if c.canClose {
		return c.Listener.Close()
	}
	return nil
}

func (c *customListener) Accept() (net.Conn, error) {
	go c.doAccept()
	select {
	case <-c.stop:
		return nil, errors.New("can not accpet on this connection")
	case connection := <-c.accept:
		go func() {
			// wait for timeout after stop and close all active connections
			<-c.stop
			time.Sleep(c.timeout)
			connection.conn.Close()
		}()
		return connection.conn, connection.err
	}
}

func (c *customListener) doAccept() {
	conn, err := c.Listener.Accept()
	c.accept <- &acceptValues{
		conn: conn,
		err:  err,
	}
}

func (c *customListener) CanClose(canClose bool) {
	c.canClose = canClose
}

func (c *customListener) GetListener() CustomListener {
	return newListener(c.Listener, c.accept, c.timeout)
}

func (c *customListener) StopAccept() {
	select {
	case _, open := <-c.stop:
		if open {
			close(c.stop)
		}
	default:
		close(c.stop)
	}
}

//NewListener creates a new CustomListener
func NewListener(network, laddr string) (CustomListener, error) {
	return NewListenerWithTimeout(network, laddr, time.Second)
}

func NewListenerWithTimeout(network, laddr string, timeout time.Duration) (CustomListener, error) {
	lis, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return newListener(lis, make(chan *acceptValues, 0), timeout), nil
}

func newListener(lis net.Listener, accept chan *acceptValues, timeout time.Duration) CustomListener {
	if timeout < time.Millisecond {
		timeout = time.Millisecond * 100
	}
	l := &customListener{
		Listener: lis,
		canClose: false,
		accept:   accept,
		stop:     make(chan *bool, 0),
		timeout:  timeout,
	}
	return l
}
