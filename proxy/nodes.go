package proxy

import (
	"net"
)

// Client implements the client-side of a tunnel.
type Client interface {
	ProxyConnection(Endpoint, net.Conn) error
}

// Server implements the server-side of a tunnel.
type Server interface {
	Start() error
}
