package https

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
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
	previous  chan *response
}

type response struct {
	reqID    string
	respData []proxy.Packet
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
		cmd:      make(chan string),
		cmdDone:  make(chan struct{}),
		proxy:    proxy.NewProxy(),
		buffer:   make([]proxy.Packet, serverBufferSize),
		previous: make(chan *response, 1),
	}

	s.previous <- &response{
		reqID:    "",
		respData: []proxy.Packet{},
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

//go:embed static/*
var staticFiles embed.FS
var staticHandler = func() http.Handler {
	fSys, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(fSys))
}()

// authenticate request
func handler(w http.ResponseWriter, r *http.Request) {
	if s.authenticate(r.Header.Get("Auth-Token")) {
		proxyHandler(w, r) // authenticated proxy handler
	} else {
		staticHandler.ServeHTTP(w, r) // decoy handler
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "error: method must be POST", http.StatusMethodNotAllowed)
		return
	}

	id := r.Header.Get("ID")
	if id == "" {
		http.Error(w, "error: ID header must be included", http.StatusBadRequest)
		return
	}

	reqBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error while reading client payload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	var packets []proxy.Packet
	if err := json.Unmarshal(reqBytes, &packets); err != nil {
		http.Error(w, "error: failed to parse JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	prev := <-s.previous

	if id != prev.reqID { // previous request was successful
		go s.proxy.Ingest(packets)

		prev.reqID = id
		prev.respData = s.buffer[:s.proxy.Fill(s.buffer)]
	}

	s.previous <- prev

	payload, err := json.Marshal(prev.respData)
	if err != nil {
		http.Error(w, "error: failed to marshal return payload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(payload); err != nil {
		http.Error(w, "error: failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
