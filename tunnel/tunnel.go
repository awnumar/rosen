// package tunnel is a simple wrapper over tunnel/wrapper that allows callers to easily communicate []router.Packet slices over arbitrary io.ReadWriters
package tunnel

import (
	"encoding/gob"
	"io"

	"github.com/awnumar/rosen/router"
	"github.com/awnumar/rosen/tunnel/wrapper"
)

type Tunnel struct {
	send *gob.Encoder
	recv *gob.Decoder
}

func New(conn io.ReadWriter, key []byte) (*Tunnel, error) {
	wrapper, err := wrapper.New(conn, key)
	if err != nil {
		return nil, err
	}
	return &Tunnel{
		send: gob.NewEncoder(wrapper),
		recv: gob.NewDecoder(wrapper),
	}, nil
}

func (t *Tunnel) Send(data []router.Packet) error {
	return t.send.Encode(data)
}

func (t *Tunnel) Recv() ([]router.Packet, error) {
	var packets []router.Packet
	return packets, t.recv.Decode(&packets)
}
