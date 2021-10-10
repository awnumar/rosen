package transport

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lukechampine.com/frand"
)

func TestSecureConnReadWrite(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{})
	require.NoError(t, err)
	defer listener.Close()

	testReadWritePayload(t, listener)
	testReadWriteSecureConn(t, listener)
}

func testReadWritePayload(t *testing.T, listener *net.TCPListener) {
	A, B := setupLocalConn(t, listener)
	defer func() {
		A.Close()
		B.Close()
	}()

}

func testReadWriteSecureConn(t *testing.T, listener *net.TCPListener) {
	A, B := setupLocalConn(t, listener)
	defer func() {
		A.Close()
		B.Close()
	}()

	secureConnA, err := SecureConnection(A)
	require.NoError(t, err)
	secureConnB, err := SecureConnection(B)
	require.NoError(t, err)

	var dataRef []byte

	for i := 0; i < 4; i++ {
		data := frand.Bytes(frand.Intn(10000))
		dataRef = append(dataRef, data...)
		n, err := secureConnA.Write(data)
		require.NoError(t, err)
		assert.Equal(t, len(data), n)
	}

	readData := make([]byte, len(dataRef))
	n, err := io.ReadFull(secureConnB, readData)
	require.NoError(t, err)
	assert.Equal(t, len(readData), n)
	assert.Equal(t, dataRef, readData)
}

func setupLocalConn(t *testing.T, listener *net.TCPListener) (serverConn, clientConn *net.TCPConn) {
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

	clientConn, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	require.NoError(t, err)

	select {
	case err := <-errChannel:
		require.NoError(t, err)
		return
	case serverConn := <-connChannel:
		return serverConn, clientConn
	}
}
