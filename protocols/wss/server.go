package wss

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/awnumar/rosen/config"
)

type Server struct{}

var s *Server

const bufferSize = 4096

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufferSize,
	WriteBufferSize: bufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewServer(conf config.Configuration) (*Server, error) {
	return &Server{}, nil
}

func (s *Server) Start() error {
	return http.ListenAndServe("127.0.0.1:23579", http.HandlerFunc(handler))
}

func handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	conn := ws.UnderlyingConn()

}
