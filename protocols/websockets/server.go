package wss

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/awnumar/rosen/lib"
	"github.com/awnumar/rosen/lib/config"
	"github.com/awnumar/rosen/protocols/https"

	"github.com/gorilla/websocket"
)

const bufferSize = 4096

type Server struct {
	s    *https.Server
	conf config.Configuration
}

var s *Server

func NewServer(conf config.Configuration) (*Server, error) {
	s_, err := https.NewServerWithCustomHandlers(conf, handler, nil)
	if err != nil {
		return nil, err
	}

	s = &Server{s: s_, conf: conf}

	return s, nil
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

	for {
		messageType, r, err := ws.NextReader()
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Println(messageType)

		if messageType != websocket.BinaryMessage {
			continue
		}

		var b []byte

		g := gob.NewDecoder(r)

		if err := g.DecodeValue(reflect.ValueOf(b)); err != nil {
			fmt.Println(err)
			break
		}

		key, err := key()
		if err != nil {
			fmt.Println(err)
			continue
		}

		plaintext, err := lib.Ae_decrypt(b, key)
		if err != nil {
			fmt.Println(err)
			continue
		}

	}
}

func key() ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s.conf["authToken"])
}
