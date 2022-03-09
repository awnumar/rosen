package tcp

import (
	"fmt"
	"net"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/awnumar/rosen/config"
	"github.com/awnumar/rosen/router"
	"github.com/awnumar/rosen/tunnel"
)

type Server struct {
	router *router.Router
	key    []byte
	port   int
}

type Client struct {
	router     *router.Router
	key        []byte
	remoteAddr *net.TCPAddr
	remoteConn *net.TCPConn
}

func NewServer(conf config.Configuration) (*Server, error) {
	key, err := config.DecodeKeyString(conf["authToken"])
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(conf["serverPort"])
	if err != nil {
		return nil, err
	}
	return &Server{
		router: router.NewRouter(),
		key:    key,
		port:   port,
	}, nil
}

func NewClient(conf config.Configuration) (*Client, error) {
	key, err := config.DecodeKeyString(conf["authToken"])
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(conf["serverPort"])
	if err != nil {
		return nil, err
	}

	var serverAddrs []net.IP
	serverAddr := conf["serverAddr"]
	if !govalidator.IsIP(serverAddr) {
		// assume serverAddr is a DNS name
		ips, err := net.LookupIP(serverAddr)
		if err != nil {
			return nil, fmt.Errorf("error: failed to lookup IP for %s: %s", serverAddr, err)
		}
		serverAddrs = ips
	} else {
		serverAddrs = []net.IP{net.ParseIP(serverAddr)}
	}

	remoteAddr := &net.TCPAddr{
		IP:   serverAddrs[0],
		Port: port,
	}

	r := router.NewRouter()

	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	go func(conn *net.TCPConn) {
		tunnel, err := tunnel.New(conn, key)
		if err != nil {
			// todo: redial and retry; handle
			fmt.Println("error creating tunnel:", err)
		}
		fmt.Println("exiting tunnel.proxywithrouter:", tunnel.ProxyWithRouter(r))
		// todo: redial and retry
	}(conn)

	return &Client{
		router:     r,
		key:        key,
		remoteAddr: remoteAddr,
		remoteConn: conn,
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: s.port,
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
			tunnel, err := tunnel.New(conn, s.key)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(tunnel.ProxyWithRouter(s.router))
		}(conn)
	}
}

func (c *Client) HandleConnection(dest router.Endpoint, conn net.Conn) error {
	return c.router.HandleConnection(dest, conn)
}
