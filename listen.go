package proxyprotocol

import (
	"net"
)

type listener struct {
	raw net.Listener
}

func Listen(raw net.Listener) net.Listener {
	return &listener{raw}
}

func (l *listener) Accept() (c net.Conn, err error) {
	raw, err := l.raw.Accept()
	if err == nil {
		c = &conn{
			raw: raw,
		}
	}
	return
}

func (l *listener) Close() error {
	return l.raw.Close()
}

func (l *listener) Addr() net.Addr {
	return l.raw.Addr()
}
