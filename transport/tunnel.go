// package transport implements an end to end encrypted tunnel over arbitrary connected reader-writer interfaces.
package transport

import (
	"encoding/gob"
	"io"

	"github.com/awnumar/rosen/router"
)

type Tunnel struct {
	send *gob.Encoder
	recv *gob.Decoder
}

func NewTunnel(conn io.ReadWriter) (*Tunnel, error) {
	secureConn, err := SecureConnection(conn)
	if err != nil {
		return nil, err
	}
	return &Tunnel{
		send: gob.NewEncoder(secureConn),
		recv: gob.NewDecoder(secureConn),
	}, nil
}

func (t *Tunnel) Send(data []router.Packet) error {
	return t.send.Encode(data)
}

func (t *Tunnel) Recv() ([]router.Packet, error) {
	var packets []router.Packet
	return packets, t.recv.Decode(&packets)
}
