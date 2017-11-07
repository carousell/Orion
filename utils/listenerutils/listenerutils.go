package listenerutils

import (
	"errors"
	"net"
)

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
	return newListener(c.Listener, c.accept)
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

func NewListener(network, laddr string) (CustomListener, error) {
	lis, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return newListener(lis, make(chan *acceptValues, 0)), nil
}

func newListener(lis net.Listener, accept chan *acceptValues) CustomListener {
	return &customListener{
		Listener: lis,
		canClose: false,
		accept:   accept,
		stop:     make(chan *bool, 0),
	}
}
