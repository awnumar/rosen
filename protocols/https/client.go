package https

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/awnumar/rosen/protocols/config"
	"github.com/awnumar/rosen/proxy"

	"lukechampine.com/frand"
)

// C implements a HTTPS tunnel client.
type C struct {
	remote    string
	transport *http.Transport
	proxy     *proxy.Proxy
}

// NewClient returns a new HTTPS client.
func NewClient(remote string, conf config.Configuration) (*C, error) {
	if !strings.HasPrefix(remote, "https://") {
		return nil, errors.New("remote address must start with https://")
	}

	trustPool, err := getTrustedCerts(conf)
	if err != nil {
		return nil, err
	}

	c := &C{
		remote: remote,
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: trustPool,
			},
		},
		proxy: proxy.NewProxy(),
	}

	go func(c *C) {
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
func (c *C) ProxyConnection(network, address string, conn net.Conn) error {
	return c.proxy.ProxyConnection(network, address, conn)
}

func getTrustedCerts(conf config.Configuration) (*x509.CertPool, error) {
	if conf["tlsCert"] == "" {
		return nil, errors.New("TLS certificate must be specified")
	}

	trustPool := x509.NewCertPool()

	pemData, err := base64.RawStdEncoding.DecodeString(conf["tlsCert"])
	if err != nil {
		return nil, err
	}

	if ok := trustPool.AppendCertsFromPEM(pemData); !ok {
		return nil, errors.New("could not parse TLS certificate")
	}

	return trustPool, nil
}
