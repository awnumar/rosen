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

	sendErrChan := make(chan error) // to client
	go func() {
		buffer := make([]router.Packet, bufferSize)
		for {
			size := s.router.Fill(buffer)
			if err := tunnel.Send(buffer[:size]); err != nil {
				sendErrChan <- err
				return
			}
		}
	}()

	recvErrChan := make(chan error) // from client
	go func() {
		for {
			data, err := tunnel.Recv()
			if err != nil {
				recvErrChan <- err
				return
			}
			s.router.Ingest(data)
		}
	}()

	select {
	case err := <-sendErrChan:
		close(sendErrChan)
		close(recvErrChan)
		log.Println(err)
		return
	case err := <-recvErrChan:
		close(sendErrChan)
		close(recvErrChan)
		log.Println(err)
		return
	}
}
