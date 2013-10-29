package proxyprotocol

import (
	"bufio"
	"bytes"
	parser "github.com/racker/go-proxy-protocol"
	"io"
	"net"
	"time"
)

type conn struct {
	raw   net.Conn
	proxy *parser.ProxyLine
	err   error
}

func (c *conn) init() {
	if c.proxy != nil || c.err != nil {
		return
	}

	// prepare for the longest (TCP6) line
	buf := make([]byte, 107)

	// read at least the shortest (UNKNOWN) line
	n, err := c.raw.Read(buf[:15])
	if err != nil && !(err == io.EOF && n == 15) {
		c.err = err
		return
	}
	if n < 15 {
		c.err = io.EOF
		return
	}

	for i := n; i < len(buf); i++ {
		if buf[i-2] == byte('\r') {
			if buf[i-1] == byte('\n') {
				buf = buf[:i]
				break
			} else {
				c.err = parser.InvalidProxyLine
				return
			}
		}

		n, c.err = c.raw.Read(buf[i : i+1])
		if c.err != nil {
			return
		}
		if n != 1 {
			c.err = io.EOF
			return
		}
	}

	if bytes.HasPrefix(buf, []byte("PROXY UNKNOWN")) {
		c.proxy = new(parser.ProxyLine)
		return
	}

	c.proxy, c.err = parser.ConsumeProxyLine(bufio.NewReader(bytes.NewReader(buf)))
	return
}

func (c *conn) Read(b []byte) (n int, err error) {
	c.init()
	if c.err != nil {
		err = c.err
	} else {
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
	c.init()
	if c.err == nil && c.proxy.DstAddr != nil {
		return c.proxy.DstAddr
	}
	return c.raw.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	c.init()
	if c.err == nil && c.proxy.SrcAddr != nil {
		return c.proxy.SrcAddr
	}
	return c.raw.RemoteAddr()
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
