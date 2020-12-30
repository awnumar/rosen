package proxy

import (
	"encoding/base64"
	"net"
	"sync"
	"sync/atomic"

	"lukechampine.com/frand"
)

const (
	fromConnsChannelBufferSize = 4096
	toConnChannelBufferSize    = 4096
)

// Proxy is a black-box structure that will proxy data between the caller and multiple connections.
type Proxy struct {
	fromConns chan Packet
	handlers  *sync.Map // string => *pipe
}

type pipe struct {
	toConn chan Packet
	close  uint32
}

// NewProxy initialises a new Proxy object.
func NewProxy() *Proxy {
	return &Proxy{
		fromConns: make(chan Packet, fromConnsChannelBufferSize),
		handlers:  &sync.Map{},
	}
}

// ProxyConnection will start handlers for a connection that wishes to talk to a given endpoint.
// If conn == nil, a connection to the given endpoint will be opened.
// Otherwise, a packet containing instructions to open a connection is sent on the p.fromConns channel.
func (p *Proxy) ProxyConnection(dest Endpoint, conn net.Conn) (err error) {
	id := base64.RawStdEncoding.EncodeToString(frand.Bytes(16))
	return p.proxyConnection(id, dest, conn)
}

func (p *Proxy) proxyConnection(id string, dest Endpoint, conn net.Conn) (err error) {
	if conn == nil {
		conn, err = net.Dial(dest.Network, dest.Address)
		if err != nil {
			return err
		}
	} else {
		p.fromConns <- NewPacket(id, dest)
	}

	toConn := make(chan Packet, toConnChannelBufferSize)
	pipe := &pipe{toConn: toConn}
	p.handlers.Store(id, pipe)

	go func() {
		for message := range toConn {
			// sanity check
			if message.ID != id {
				panic("received a message intended for a different client; please report this issue")
			}

			if message.Closed() {
				break
			}

			_, err = conn.Write(message.Data)
			if err != nil {
				p.fromConns <- ClosePacket(message.ID)
				break
			}
		}
		atomic.StoreUint32(&pipe.close, 1)
		conn.Close()
	}()

	go func() {
		copyBuf := func(buf []byte) []byte {
			c := make([]byte, len(buf))
			copy(c, buf)
			return c
		}
		var readBuf [4000000]byte
		for {
			n, err := conn.Read(readBuf[:])
			if n > 0 {
				p.fromConns <- DataPacket(id, copyBuf(readBuf[:n]))
			}
			if err != nil {
				p.fromConns <- ClosePacket(id)
				break
			}
		}
		atomic.StoreUint32(&pipe.close, 1)
		conn.Close()
	}()

	return nil
}

// Ingest takes a list of packets and handles them, forwarding data to the right handlers.
func (p *Proxy) Ingest(data []Packet) {
	for i := range data {
		id := data[i].ID

		pipeInterface, exists := p.handlers.Load(id)
		if !exists {
			if data[i].NewConnection() {
				p.proxyConnection(data[i].ID, data[i].Dest, nil)
			}
			continue
		}
		pipe := pipeInterface.(*pipe) // will panic if can't assert type

		if atomic.LoadUint32(&pipe.close) == 1 {
			p.handlers.Delete(id)
			close(pipe.toConn)
			continue
		}

		pipe.toConn <- data[i]
	}
}

// QueueLen returns the number of packets waiting on the aggregate channel of data from connections.
func (p *Proxy) QueueLen() int {
	return len(p.fromConns)
}

// Fill tries to fill provided buffer with waiting items from the connections' outbound queue.
// It returns the number of packets written.
func (p *Proxy) Fill(buffer []Packet) int {
	size := p.QueueLen()
	if size > len(buffer) {
		size = len(buffer)
	}
	for i := 0; i < size; i++ {
		buffer[i] = <-p.fromConns
	}
	return size
}
