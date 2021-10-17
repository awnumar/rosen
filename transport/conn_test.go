package transport

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lukechampine.com/frand"
)

func TestReadWriteSecureConn(t *testing.T) {
	A, B, err := setupLocalConn()
	require.NoError(t, err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	sA, err := SecureConnection(A, key)
	require.NoError(t, err)
	sB, err := SecureConnection(B, key)
	require.NoError(t, err)

	refData := make([]byte, 0)

	for i := 0; i < 4; i++ {
		data := frand.Bytes(frand.Intn(4096))
		_, err = sA.Write(data)
		require.NoError(t, err)
		refData = append(refData, data...)
	}

	readData := make([]byte, len(refData))
	_, err = io.ReadFull(sB, readData)
	require.NoError(t, err)

	require.True(t, bytes.Equal(refData, readData))
}

func TestReadWritePayload(t *testing.T) {
	A, B, err := setupLocalConn()
	require.NoError(t, err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	sA, err := SecureConnection(A, key)
	require.NoError(t, err)
	sB, err := SecureConnection(B, key)
	require.NoError(t, err)

	refData := make([]byte, 0)

	for i := 0; i < 4; i++ {
		data := frand.Bytes(frand.Intn(4096))
		require.NoError(t, sA.writePayload(data))
		refData = append(refData, data...)
	}

	readData := make([]byte, 0)
	for i := 0; i < 4; i++ {
		data, err := sB.readPayload()
		require.NoError(t, err)
		readData = append(readData, data...)
	}

	assert.Equal(t, refData, readData)
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
