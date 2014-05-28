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
	buf   []byte
	err   error
	proxy *parser.ProxyLine
}

func (c *conn) init() {
	if c.buf != nil || c.err != nil || c.proxy != nil {
		return
	}

	// prepare for the longest (TCP6) line
	buf := make([]byte, 107)

	// read at least the shortest (UNKNOWN) line
	n, err := c.raw.Read(buf[:15])
	if err != nil && !(err == io.EOF && n == 15) {
		c.buf = buf[:n]
		c.err = err
		return
	}
	if n < 15 {
		c.buf = buf[:n]
		c.err = io.EOF
		return
	}

	for i := n; i < len(buf); i++ {
		if buf[i-2] == byte('\r') {
			if buf[i-1] == byte('\n') {
				buf = buf[:i]
				break
			} else {
				c.buf = buf[:i]
				return
			}
		}

		n, c.err = c.raw.Read(buf[i : i+1])
		if c.err != nil {
			c.buf = buf[:i+n]
			return
		}
		if n == 0 {
			c.buf = buf[:i]
			c.err = io.EOF
			return
		}
	}

	if bytes.HasPrefix(buf, []byte("PROXY UNKNOWN")) {
		c.proxy = new(parser.ProxyLine)
		return
	}

	c.proxy, c.err = parser.ConsumeProxyLine(bufio.NewReader(bytes.NewReader(buf)))
	if c.err == nil && c.proxy == nil {
		c.buf = buf
	}
	return
}

func (c *conn) Read(b []byte) (n int, err error) {
	c.init()

	if len(c.buf) > 0 {
		n = copy(b, c.buf)
		b = b[n:]
		c.buf = c.buf[:n]

		if len(b) == 0 {
			return
		}
	}

	if c.err != nil {
		err = c.err
		return
	}

	var l int

	l, err = c.raw.Read(b)
	n += l

	return
}

func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.raw.Write(b)
	return
}

func (c *conn) Close() error {
	c.buf = nil
	return c.raw.Close()
}

func (c *conn) LocalAddr() net.Addr {
	c.init()

	if c.proxy != nil && c.proxy.DstAddr != nil {
		return c.proxy.DstAddr
	}

	return c.raw.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	c.init()

	if c.proxy != nil && c.proxy.SrcAddr != nil {
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
