package router

import (
	"net"
)

// Client implements the client-side of a tunnel.
type Client interface {
	HandleConnection(dest Endpoint, conn net.Conn) error
}

// Server implements the server-side of a tunnel.
type Server interface {
	Start() error
}
