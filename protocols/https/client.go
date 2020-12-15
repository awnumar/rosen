package https

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/awnumar/rosen/protocols/config"
	"github.com/awnumar/rosen/proxy"

	"github.com/asaskevich/govalidator"
	"lukechampine.com/frand"
)

// Client implements a HTTPS tunnel client.
type Client struct {
	remote    string
	transport *http.Transport
	proxy     *proxy.Proxy
}

// NewClient returns a new HTTPS client.
func NewClient(conf config.Configuration) (*Client, error) {
	if !govalidator.IsURL(conf["proxyAddr"]) {
		return nil, errors.New("config: proxy address must be an URL")
	}
	if !strings.HasPrefix(conf["proxyAddr"], "https://") {
		return nil, errors.New("config: proxy address must start with https://")
	}

	trustPool, err := trustedCertPool(conf["pinRootCA"])
	if err != nil {
		return nil, err
	}

	c := &Client{
		remote: conf["proxyAddr"],
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: trustPool,
			},
		},
		proxy: proxy.NewProxy(),
	}

	go func(c *Client) {
		var responseData []proxy.Packet
		outboundBuffer := make([]proxy.Packet, clientBufferSize)

		for {
			size := c.proxy.Fill(outboundBuffer)

			payload, err := json.Marshal(outboundBuffer[:size])
			if err != nil {
				panic("failed to encode message payload; error: " + err.Error())
			}

			req, err := http.NewRequest(http.MethodPost, c.remote, bytes.NewReader(payload))
			if err != nil {
				panic(err)
			}

			req.Header.Set("Auth-Token", conf["authToken"])

			resp, err := c.transport.RoundTrip(req) // todo: implement retries
			if err != nil {
				panic(err) // retry condition
			}

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err) // retry condition
			}
			resp.Body.Close()

			if err := json.Unmarshal(respBytes, &responseData); err != nil {
				panic(err) // maybe auth token is wrong
			}

			go c.proxy.Ingest(responseData)

			if size > 0 || c.proxy.QueueLen() > 0 || len(responseData) > 0 {
				continue
			}

			// delay randomly, sample from some distribution, or based on previous delay?
			// copy meek?
			// for now:
			// random between 0 and 100 milliseconds, with nanosecond resolution
			time.Sleep(time.Duration(frand.Intn(1_000_000_00)) * time.Nanosecond)
		}
	}(c)

	return c, nil
}

// ProxyConnection handles and proxies a single connection between a local client and the remote server.
func (c *Client) ProxyConnection(dest proxy.Endpoint, conn net.Conn) error {
	return c.proxy.ProxyConnection(dest, conn)
}
