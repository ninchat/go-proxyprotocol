package proxyprotocol

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"
)

var TODO = errors.New("TODO")

type conn struct {
	raw    net.Conn
	remote net.Addr
}

func (c *conn) init() (err error) {
	if c.remote != nil {
		return
	}

	// fallback
	c.remote = c.raw.RemoteAddr()

	buf := make([]byte, 107)

	n, err := c.raw.Read(buf[:32])
	if err != nil {
		if err == io.EOF && n == 32 {
			err = nil
		} else {
			return
		}
	}
	if n < 32 {
		err = TODO
		return
	}

	for i := n; i < len(buf); i++ {
		if buf[i-2] == byte('\r') {
			if buf[i-1] == byte('\n') {
				buf = buf[:i-2]
				break
			} else {
				err = TODO
				return
			}
		}

		n, err = c.raw.Read(buf[i:i+1])
		if err != nil {
			return
		}
		if n != 1 {
			err = TODO
			return
		}
	}

	if !bytes.HasPrefix(buf, []byte("PROXY ")) {
		err = TODO
		return
	}
	buf = buf[6:]

	var addr  addr
	var delim byte

	if bytes.HasPrefix(buf, []byte("TCP4 ")) {
		addr.network = "tcp4"
		delim = byte('.')
	} else if bytes.HasPrefix(buf, []byte("TCP6 ")) {
		addr.network = "tcp6"
		delim = byte(':')
	} else {
		if !bytes.HasPrefix(buf, []byte("UNKNOWN ")) {
			err = TODO
		}
		return
	}
	buf = buf[5:]

	i := bytes.IndexByte(buf, byte(' '))
	if i < 7 {
		err = TODO
		return
	}
	addr.address = string(buf[:i])
	if bytes.IndexByte(buf[:i], delim) < 0 || net.ParseIP(addr.address) == nil {
		err = TODO
		return
	}

	c.remote = &addr
	return
}

func (c *conn) Read(b []byte) (n int, err error) {
	err = c.init()
	if err == nil {
		n, err = c.raw.Read(b)
	}
	return
}

func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.raw.Write(b)
	return
}

func (c *conn) Close() error {
	return c.raw.Close()
}

func (c *conn) LocalAddr() net.Addr {
	return c.raw.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	c.init()
	return c.remote
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.raw.SetDeadline(t)
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.raw.SetReadDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.raw.SetWriteDeadline(t)
}
