package listenerutils

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/carousell/Orion/utils/log"
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
	stop     chan struct{}
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

type customConn struct {
	net.Conn
	closed chan struct{}
}

func (c *customConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.checkErrorAndClose(err)
	return
}

func (c *customConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	c.checkErrorAndClose(err)
	return
}

func (c *customConn) Close() error {
	c.doClose()
	return c.Conn.Close()
}

func (c *customConn) checkErrorAndClose(err error) {
	if err != nil {
		if err == io.EOF {
			c.doClose()
		} else if e, ok := err.(net.Error); ok {
			// close on non temporary error
			if !e.Temporary() {
				c.doClose()
			}
		} else {
			// this should not happen as all errors
			// returned from Read should implement net.Error
			// but its better we close on all other errors
			c.doClose()
		}
	}
}

func (c *customConn) watcher(stop chan struct{}, timeout time.Duration) {
	// wait for timeout after stop and close all active connections
	select {
	case <-stop:
		time.Sleep(timeout)
		c.Close()
	case <-c.closed:
		// do nothing connection is already closed
		return
	}
}

func (c *customConn) doClose() {
	defer func() {
		if err := recover(); err != nil {
			log.Info(context.Background(), "msg", "panic trying to close channel", "err", err)
		}
	}()
	select {
	case <-c.closed:
	// do nothing already closed
	default:
		close(c.closed)
	}
}

func (c *customListener) Accept() (net.Conn, error) {
	go c.doAccept()
	select {
	case <-c.stop:
		return nil, errors.New("can not accpet on this connection")
	case connection := <-c.accept:
		conn := &customConn{
			Conn:   connection.conn,
			closed: make(chan struct{}, 0),
		}
		go conn.watcher(c.stop, c.timeout)
		return conn, connection.err
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
	defer func() {
		if err := recover(); err != nil {
			log.Info(context.Background(), "msg", "panic trying to close channel", "err", err)
		}
	}()
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
		stop:     make(chan struct{}, 0),
		timeout:  timeout,
	}
	return l
}
