package transport

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/awnumar/rosen/crypto"
)

type SecureConn struct {
	key  []byte
	conn io.ReadWriter

	readMutex  *sync.Mutex
	writeMutex *sync.Mutex

	readBuffer chan []byte
}

func SecureConnection(conn io.ReadWriter) (*SecureConn, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return &SecureConn{
		key:        key,
		conn:       conn,
		readMutex:  &sync.Mutex{},
		writeMutex: &sync.Mutex{},
		readBuffer: make(chan []byte, 1),
	}, nil
}

func (s *SecureConn) Read(b []byte) (int, error) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()

	select {
	case buffered := <-s.readBuffer:
		_ = buffered
	default:

	}

	ciphertext, err := readPayload(s.conn)
	if err != nil {
		return 0, err
	}

	data, err := crypto.Decrypt(ciphertext, s.key)
	if err != nil {
		return 0, err
	}

	if len(b) < len(data) {
		n := copy(b, data)
		s.readBuffer <- data[n:]
		return n, nil
	}

	return copy(b, data), nil
}

func (s *SecureConn) Write(b []byte) (int, error) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	ciphertext, err := crypto.Encrypt(b, s.key)
	if err != nil {
		return 0, err
	}

	return writePayload(s.conn, ciphertext)
}

func readPayload(conn io.Reader) ([]byte, error) {
	lengthBytes := make([]byte, binary.MaxVarintLen64)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err
	}
	length, bytesRead := binary.Uvarint(lengthBytes)
	if bytesRead <= 0 {
		// an error occurred
		if bytesRead == 0 {
			// buffer was too small
			panic("buffer too small to hold length; please report this bug")
		} else {
			// value larger than 64 bits. bytes read is -bytesRead
			// todo: in the future, set an extra byte to indicate that message overflows to next transmission
			return nil, fmt.Errorf("length of buffer cannot be larger than 64 bits")
		}
	}
	if length == 0 {
		// empty message? return error for now
		return nil, fmt.Errorf("zero length message")
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	return data, nil
}

func writePayload(conn io.Writer, data []byte) (int, error) {
	lengthBytes := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(lengthBytes, uint64(len(data)))
	return conn.Write(append(lengthBytes, data...))
}
