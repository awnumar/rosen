package router

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

// Router is a black-box structure that will route data between the caller and multiple connections.
type Router struct {
	fromConns chan Packet
	handlers  *sync.Map // string => *pipe
}

type pipe struct {
	toConn chan Packet
	close  uint32
}

// NewRouter initialises a new Router object.
func NewRouter() *Router {
	return &Router{
		fromConns: make(chan Packet, fromConnsChannelBufferSize),
		handlers:  &sync.Map{},
	}
}

// RouterConnection will start handlers for a connection that wishes to talk to a given endpoint.
// If conn == nil, a connection to the given endpoint will be opened.
// Otherwise, a packet containing instructions to open a connection is sent on the p.fromConns channel.
func (r *Router) HandleConnection(dest Endpoint, conn net.Conn) (err error) {
	id := base64.RawStdEncoding.EncodeToString(frand.Bytes(16))
	return r.handleConnection(id, dest, conn)
}

func (r *Router) handleConnection(id string, dest Endpoint, conn net.Conn) (err error) {
	if conn == nil {
		conn, err = net.Dial(dest.Network, dest.Address)
		if err != nil {
			return err
		}
	} else {
		r.fromConns <- NewPacket(id, dest)
	}

	toConn := make(chan Packet, toConnChannelBufferSize)
	pipe := &pipe{toConn: toConn}
	r.handlers.Store(id, pipe)

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
				r.fromConns <- ClosePacket(message.ID)
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
				r.fromConns <- DataPacket(id, copyBuf(readBuf[:n]))
			}
			if err != nil {
				r.fromConns <- ClosePacket(id)
				break
			}
		}
		atomic.StoreUint32(&pipe.close, 1)
		conn.Close()
	}()

	return nil
}

// Ingest takes a list of packets and handles them, forwarding data to the right handlers.
func (r *Router) Ingest(data []Packet) {
	for i := range data {
		id := data[i].ID

		pipeInterface, exists := r.handlers.Load(id)
		if !exists {
			if data[i].NewConnection() {
				r.handleConnection(data[i].ID, data[i].Dest, nil)
			}
			continue
		}
		pipe := pipeInterface.(*pipe) // will panic if can't assert type

		if atomic.LoadUint32(&pipe.close) == 1 {
			r.handlers.Delete(id)
			close(pipe.toConn)
			continue
		}

		pipe.toConn <- data[i]
	}
}

// QueueLen returns the number of packets waiting on the aggregate channel of data from connections.
func (r *Router) QueueLen() int {
	return len(r.fromConns)
}

// Fill tries to fill provided buffer with waiting items from the connections' outbound queue.
// It returns the number of packets written.
func (r *Router) Fill(buffer []Packet) int {
	size := r.QueueLen()
	if size > len(buffer) {
		size = len(buffer)
	}
	for i := 0; i < size; i++ {
		buffer[i] = <-r.fromConns
	}
	return size
}
