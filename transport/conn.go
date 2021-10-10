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

	readPayloads chan *payload // holds some data chunks ready for Read
	readBuffer   []byte        // buffer Read uses if caller's buffer is too small

	readMutex  *sync.Mutex // syncs all access to readBuffer
	writeMutex *sync.Mutex // syncs writes to conn
}

type payload struct {
	data []byte
	err  error
}

const overhead = crypto.Overhead + binary.MaxVarintLen64

func SecureConnection(conn io.ReadWriter) (*SecureConn, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	readPayloads := make(chan *payload, 2)

	go func() {
		for {
			ciphertext, err := readPayload(conn)
			if err != nil {
				readPayloads <- &payload{err: err} // enhancement: reconnect on broken pipe
				close(readPayloads)
				return
			}

			data, err := crypto.Decrypt(ciphertext, key)
			if err != nil {
				fmt.Println(ciphertext)
				readPayloads <- &payload{err: err} // key is wrong or authentication failed
				close(readPayloads)
				return
			}

			// blocks if the channel is full
			readPayloads <- &payload{data: data}
		}
	}()

	return &SecureConn{
		key:          key,
		conn:         conn,
		readPayloads: readPayloads,
		readMutex:    &sync.Mutex{},
		writeMutex:   &sync.Mutex{},
	}, nil
}

func (s *SecureConn) Read(b []byte) (int, error) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()

	ptr := 0
	if len(s.readBuffer) > 0 {
		n := copy(b, s.readBuffer)
		if n == len(b) { // filled b
			// so len(b) <= len(readBuffer)
			if len(b) < len(s.readBuffer) {
				// leftover data in the buffer
				s.readBuffer = s.readBuffer[n:]
			}
			return n, nil
		}
		// n > len(b) or n < len(b)
		// n cannot be more than len(b)
		// therefore, n < len(b)
		// since n is minimum of len(b) and len(readBuffer),
		// n == len(readBuffer)
		// therefore, emptied buffer and didn't fill b
		ptr = n
	}

	select {
	case payload, ok := <-s.readPayloads:
		if payload.err != nil {
			return 0, payload.err
		}

		if !ok {
			return 0, fmt.Errorf("error: connection closed")
		}

		n := copy(b[ptr:], payload.data)
		s.readBuffer = payload.data[n:]
		return ptr + n, nil

	default:
		return 0, nil
	}
}

func (s *SecureConn) Write(b []byte) (int, error) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	ciphertext, err := crypto.Encrypt(b, s.key)
	if err != nil {
		return 0, err
	}

	n, err := writePayload(s.conn, ciphertext)
	if err != nil {
		return 0, err
	}

	return n - overhead, nil
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
	fmt.Println(length)
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
