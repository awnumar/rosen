package proxy

import (
	"net"
)

// Client implements the client-side of a tunnel.
type Client interface {
	ProxyConnection(network, address string, conn net.Conn) error
}

// Server implements the server-side of a tunnel.
type Server interface {
	Start() error
}
