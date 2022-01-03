package tcp

import (
	"fmt"
	"net"

	"github.com/awnumar/rosen/router"
	"github.com/awnumar/rosen/transport"
	"golang.org/x/crypto/blake2b"
)

const (
	port = 23579
)

var key = func() []byte {
	k := blake2b.Sum256([]byte("test"))
	return k[:]
}()

type Server struct {
	r *router.Router
}

type Client struct {
	r          *router.Router
	remoteAddr *net.TCPAddr
	remoteConn *net.TCPConn
}

func NewServer() *Server {
	return &Server{
		r: router.NewRouter(),
	}
}

func NewClient() (*Client, error) {
	r := router.NewRouter()

	remoteAddr := &net.TCPAddr{
		IP:   net.ParseIP("100.96.221.43"),
		Port: port,
	}

	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	go func(conn *net.TCPConn) {
		for {
			tunnel, err := transport.NewTunnel(conn, key)
			if err != nil {
				// todo: redial and retry; handle
				fmt.Println("error creating tunnel:", err)
				continue
			}
			fmt.Println("tunnel.proxywithrouter:", tunnel.ProxyWithRouter(r))
			// todo: redial and retry
		}
	}(conn)

	return &Client{
		r:          r,
		remoteAddr: remoteAddr,
		remoteConn: conn,
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	})
	if err != nil {
		return err
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("error while accepting connection:", err)
		}

		go func(conn *net.TCPConn) {
			tunnel, err := transport.NewTunnel(conn, key)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(tunnel.ProxyWithRouter(s.r))
		}(conn)
	}
}

func (c *Client) HandleConnection(dest router.Endpoint, conn net.Conn) error {
	return c.r.HandleConnection(dest, conn)
}
