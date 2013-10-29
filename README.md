Partial PROXY protocol version 1 (server-end) support for Go.
<http://haproxy.1wt.eu/download/1.5/doc/proxy-protocol.txt>

	package example

	import (
		"github.com/ninchat/go-proxyprotocol"
		"net"
		"net/http"
	)

	func ListenAndServe(addr string, handler http.Handler) (err error) {
		l, err := net.Listen("tcp", addr)
		if err == nil {
			err = http.Serve(proxyprotocol.Listen(l), handler)
		}
		return
	}
