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
	canClose   bool
	stopAccept bool
}

func (c *customListener) Close() error {
	if c.canClose {
		return c.Listener.Close()
	}
	return nil
}
func (c *customListener) Accept() (net.Conn, error) {
	if c.stopAccept {
		return nil, errors.New("can not accpet on this connection")
	}
	return c.Listener.Accept()
}

func (c *customListener) CanClose(canClose bool) {
	c.canClose = canClose
}

func (c *customListener) GetListener() CustomListener {
	return &customListener{
		Listener: c.Listener,
	}
}

func (c *customListener) StopAccept() {
	c.stopAccept = true
}

func NewListener(network, laddr string) (CustomListener, error) {
	lis, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return &customListener{
		Listener:   lis,
		canClose:   false,
		stopAccept: false,
	}, nil
}
