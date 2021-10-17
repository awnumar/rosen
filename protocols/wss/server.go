package wss

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/awnumar/rosen/config"
	"github.com/awnumar/rosen/protocols/https"
	"github.com/awnumar/rosen/router"
	"github.com/awnumar/rosen/transport"
)

const bufferSize = 4096

type Server struct {
	key    []byte
	conf   config.Configuration
	s      *https.Server
	router *router.Router
}

var s *Server

func NewServer(conf config.Configuration) (*Server, error) {
	s, err := https.NewServerWithCustomHandlers(conf, handler, nil)
	if err != nil {
		return nil, err
	}

	key, err := config.DecodeKeyString(conf["authToken"])
	if err != nil {
		return nil, err
	}

	return &Server{
		key:    key,
		conf:   conf,
		s:      s,
		router: router.NewRouter(),
	}, nil
}

func (s *Server) Start() error {
	return s.s.Start()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufferSize,
	WriteBufferSize: bufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	tunnel, err := transport.NewTunnel(ws.UnderlyingConn(), s.key)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(tunnel.ProxyWithRouter(s.router).Error())
}
