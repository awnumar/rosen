package https

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/awnumar/rosen/protocols/config"
	"github.com/awnumar/rosen/proxy"
)

// Server implements a HTTP tunnel server.
type Server struct {
	conf      config.Configuration
	tlsConfig *tls.Config
	redirect  *http.Server
	server    *http.Server
	cmd       chan string
	cmdDone   chan struct{}
	proxy     *proxy.Proxy
	buffer    []proxy.Packet
}

var s = &Server{}

// NewServer returns a new HTTPS server.
func NewServer(conf config.Configuration) (*Server, error) {
	certReloader, err := getCertificate(conf["hostname"], conf["email"])
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

	s = &Server{
		conf: conf,
		tlsConfig: &tls.Config{
			MinVersion:       tls.VersionTLS12,
			MaxVersion:       tlsMaxVersion,
			CurvePreferences: []tls.CurveID{tls.X25519},
			GetCertificate:   certReloader.GetCertificateFunc(),
		},
		proxy:   proxy.NewProxy(),
		cmd:     make(chan string),
		cmdDone: make(chan struct{}),
		buffer:  make([]proxy.Packet, serverBufferSize),
	}

	return s, nil
}

// Start launches the server.
func (s *Server) Start() error {
	httpError := make(chan error)
	httpsError := make(chan error)
	defer close(httpError)
	defer close(httpsError)

	start := func() struct{} {
		s.redirect = &http.Server{
			Addr: ":80",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			}),
		}
		s.server = &http.Server{
			Addr:      ":443",
			Handler:   http.HandlerFunc(handler),
			TLSConfig: s.tlsConfig,
		}
		go func() {
			httpError <- s.redirect.ListenAndServe()
		}()
		go func() {
			httpsError <- s.server.ListenAndServeTLS("", "")
		}()
		return struct{}{}
	}

	shutdown := func(serv *http.Server) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		serv.Shutdown(ctx)
		cancel()
	}

	stop := func() struct{} {
		shutdown(s.server)
		shutdown(s.redirect)
		return struct{}{}
	}

	start()

	cmdShutdown := false
	for {
		select {
		case err := <-httpError:
			if !cmdShutdown {
				shutdown(s.server)
				return err
			}
		case err := <-httpsError:
			if !cmdShutdown {
				shutdown(s.redirect)
				return err
			}
		case cmd := <-s.cmd:
			switch cmd {
			case "stop":
				cmdShutdown = true
				s.cmdDone <- stop()
			case "start":
				cmdShutdown = false
				s.cmdDone <- start()
			case "end":
				cmdShutdown = true
				stop()
				return http.ErrServerClosed
			default:
				panic("error: unknown command sent to server handler; please report this issue")
			}
		}
	}
}

// Compare key with execution time that is a function of input length and not of input contents.
// Average time Delta between a valid and invalid key length is 29ns, on a Ryzen 3700X.
func (s *Server) authenticate(provided string) bool {
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
