package tunnel

import (
	"encoding/base64"
	"net"
	"testing"

	"github.com/matryer/is"
	"lukechampine.com/frand"

	"github.com/awnumar/rosen/router"
)

func TestReadWriteTunnel(t *testing.T) {
	is := is.New(t)

	A, B, err := setupLocalConn()
	is.NoErr(err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	tA, err := New(A, key)
	is.NoErr(err)
	tB, err := New(B, key)
	is.NoErr(err)

	refData := randomPacketSeq(100)

	is.NoErr(tA.Send(refData))

	readData, err := tB.Recv()
	is.NoErr(err)

	// ugly hack since assert doesn't consider nil and len(0) to be equal
	for i := range readData {
		if len(readData[i].Data) == 0 {
			readData[i].Data = nil
		}
	}

	is.Equal(refData, readData)
}

func randomPacketSeq(length int) (packets []router.Packet) {
	for i := 0; i < length; i++ {
		packets = append(packets, randomPacket())
	}
	return
}

func randomPacket() router.Packet {
	return router.Packet{
		ID: base64.RawURLEncoding.EncodeToString(frand.Bytes(16)),
		Dest: router.Endpoint{
			Network: base64.RawURLEncoding.EncodeToString(frand.Bytes(16)),
			Address: base64.RawURLEncoding.EncodeToString(frand.Bytes(16)),
		},
		Data: frand.Bytes(frand.Intn(4096)),
		Type: router.Data,
	}
}

func setupLocalConn() (*net.TCPConn, *net.TCPConn, error) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{})
	if err != nil {
		return nil, nil, err
	}
	defer listener.Close()

	connChannel := make(chan *net.TCPConn)
	errChannel := make(chan error)
	defer close(connChannel)
	defer close(errChannel)

	go func() {
		conn, err := listener.AcceptTCP()
		if err != nil {
			errChannel <- err
		} else {
			connChannel <- conn
		}
	}()

	A, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		return nil, nil, err
	}

	select {
	case err := <-errChannel:
		return nil, nil, err
	case B := <-connChannel:
		return A, B, nil
	}
}
