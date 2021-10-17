package wss

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/gorilla/websocket"

	"github.com/awnumar/rosen/config"
	"github.com/awnumar/rosen/crypto"
	"github.com/awnumar/rosen/router"
	"github.com/awnumar/rosen/transport"
)

type Client struct {
	r         *router.Router
	wssClient *websocket.Dialer
}

func NewClient(conf config.Configuration) (*Client, error) {
	trustPool, err := crypto.TrustedCertPool(conf["pinRootCA"])
	if err != nil {
		return nil, err
	}

	c := &Client{
		wssClient: &websocket.Dialer{
			TLSClientConfig: &tls.Config{
				RootCAs: trustPool,
			},
		},
		r: router.NewRouter(),
	}

	wssConn, _, err := c.wssClient.Dial(conf["proxyAddr"], nil)
	if err != nil {
		return nil, err
	}

	key, err := config.DecodeKeyString(conf["authToken"])
	if err != nil {
		return nil, err
	}

	tunnel, err := transport.NewTunnel(wssConn.UnderlyingConn(), key)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := tunnel.ProxyWithRouter(c.r); err != nil {
			log.Println(err)
			return
		}

	}()

	return c, nil
}

func (c *Client) HandleConnection(dest router.Endpoint, conn net.Conn) error {
	return c.r.HandleConnection(dest, conn)
}
