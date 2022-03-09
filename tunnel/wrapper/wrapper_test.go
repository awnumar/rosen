package wrapper

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/matryer/is"
	"lukechampine.com/frand"
)

func TestReadWriteWrapper(t *testing.T) {
	is := is.New(t)

	A, B, err := setupLocalConn()
	is.NoErr(err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	sA, err := New(A, key)
	is.NoErr(err)
	sB, err := New(B, key)
	is.NoErr(err)

	refData := make([]byte, 0)

	for i := 0; i < 4; i++ {
		data := frand.Bytes(frand.Intn(4096))
		_, err = sA.Write(data)
		is.NoErr(err)
		refData = append(refData, data...)
	}

	readData := make([]byte, len(refData))
	_, err = io.ReadFull(sB, readData)
	is.NoErr(err)

	is.True(bytes.Equal(refData, readData))
}

func TestReadWritePayload(t *testing.T) {
	is := is.New(t)

	A, B, err := setupLocalConn()
	is.NoErr(err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	sA, err := New(A, key)
	is.NoErr(err)
	sB, err := New(B, key)
	is.NoErr(err)

	refData := make([]byte, 0)

	for i := 0; i < 4; i++ {
		data := frand.Bytes(frand.Intn(4096))
		is.NoErr(sA.writePayload(data))
		refData = append(refData, data...)
	}

	readData := make([]byte, 0)
	for i := 0; i < 4; i++ {
		data, err := sB.readPayload()
		is.NoErr(err)
		readData = append(readData, data...)
	}

	is.Equal(refData, readData)
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
