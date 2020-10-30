package https

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/awnumar/rosen/protocols/config"
	"github.com/awnumar/rosen/proxy"
)

// S implements a HTTP tunnel server.
type S struct {
	conf   config.Configuration
	server *http.Server
	proxy  *proxy.Proxy
	buffer []proxy.Packet
}

var s = &S{}

// NewServer returns a new HTTPS server.
func NewServer(conf config.Configuration) (*S, error) {
	if conf["tlsCert"] == "" || conf["tlsKey"] == "" {
		return nil, errors.New("TLS certificate and key must be specified")
	}

	tlsCertPem, err := base64.RawStdEncoding.DecodeString(conf["tlsCert"])
	if err != nil {
		return nil, err
	}

	tlsKeyPem, err := base64.RawStdEncoding.DecodeString(conf["tlsKey"])
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(tlsCertPem, tlsKeyPem)
	if err != nil {
		return nil, err
	}

	var tlsMaxVersion uint16
	switch conf["tlsMaxVersion"] {
	case "1.2":
		tlsMaxVersion = tls.VersionTLS12
	case "1.3":
		tlsMaxVersion = tls.VersionTLS13
	default:
		return nil, errors.New("tlsMaxversion must be one of 1.2 or 1.3")
	}

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		MaxVersion:       tlsMaxVersion,
		CurvePreferences: []tls.CurveID{tls.X25519},
		Certificates:     []tls.Certificate{cert},
	}

	server := &http.Server{
		Addr:      ":443",
		Handler:   http.HandlerFunc(handler),
		TLSConfig: tlsConfig,
	}

	s = &S{
		conf:   conf,
		server: server,
		proxy:  proxy.NewProxy(),
		buffer: make([]proxy.Packet, serverBufferSize),
	}

	return s, nil
}

// Compare key with execution time that is a function of input length and not of input contents.
// Average time Delta between a valid and invalid key length is 29ns, on a Ryzen 3700X.
func (s *S) authenticate(provided string) bool {
	authToken := s.conf["authToken"]

	if len(provided) != len(authToken) {
		return false
	}

	equal := true
	for i := 0; i < len(authToken); i++ {
		if subtle.ConstantTimeByteEq(provided[i], authToken[i]) != 1 {
			equal = false
		}
	}

	return equal
}

// Start launches the server.
func (s *S) Start() error {
	httpError := make(chan error)
	httpsError := make(chan error)
	defer close(httpError)
	defer close(httpsError)

	httpServer := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		}),
	}

	go func() {
		httpError <- httpServer.ListenAndServe()
	}()

	go func() {
		httpsError <- s.server.ListenAndServeTLS("", "")
	}()

	select {
	case err := <-httpError:
		shutdown(s.server)
		return err
	case err := <-httpsError:
		shutdown(httpServer)
		return err
	}
}

func shutdown(s *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	s.Shutdown(ctx)
	cancel()
}

// authenticate request
func handler(w http.ResponseWriter, r *http.Request) {
	if s.authenticate(r.Header.Get("Auth-Token")) {
		proxyHandler(w, r) // authenticated proxy handler
	} else {
		staticHandler(w, r) // decoy handler
	}
}

var staticWebsiteHandler = http.FileServer(http.Dir("public"))

func staticHandler(w http.ResponseWriter, r *http.Request) {
	staticWebsiteHandler.ServeHTTP(w, r)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method must be POST", http.StatusMethodNotAllowed)
	}

	reqBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	r.Body.Close()

	var packets []proxy.Packet
	if err := json.Unmarshal(reqBytes, &packets); err != nil {
		http.Error(w, "Failed to unmarshal payload into []Message: "+err.Error(), http.StatusBadRequest)
	}

	go s.proxy.Ingest(packets)

	size := s.proxy.Fill(s.buffer)

	payload, err := json.Marshal(s.buffer[:size])
	if err != nil {
		// todo: handle gracefully
		http.Error(w, "Failed to marshal return payload: "+err.Error(), http.StatusInternalServerError)
	}

	if _, err := w.Write(payload); err != nil {
		http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
	}

	// reliability:
	// server sends back OK status on response. If client doesn't receive this then client retries request.
	// Client sends request ID with each request, server saves responses to these IDs. If client retries saved
	// request, server replays it. if client indicates id success, server deletes response data.
}
