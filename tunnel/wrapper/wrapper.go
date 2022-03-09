// package wrapper implements a wrapper for an io.ReadWriter that encrypts and authenticates data before exchanging data with the underlying io.ReadWriter
package wrapper

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/awnumar/rosen/crypto"
)

type Wrapper struct {
	conn   io.ReadWriter
	cipher *crypto.Cipher

	readBuffer []byte

	readMutex  *sync.Mutex
	writeMutex *sync.Mutex
}

func New(conn io.ReadWriter, key []byte) (*Wrapper, error) {
	cipher, err := crypto.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &Wrapper{
		conn:       conn,
		cipher:     cipher,
		readMutex:  &sync.Mutex{},
		writeMutex: &sync.Mutex{},
	}, nil
}

func (s *Wrapper) Read(b []byte) (int, error) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()

	if len(s.readBuffer) > 0 {
		n := copy(b, s.readBuffer)
		if n == len(s.readBuffer) {
			// emptied buffer
			s.readBuffer = nil
		} else {
			// did not empty buffer
			s.readBuffer = s.readBuffer[n:]
		}
		return n, nil
	}

	data, err := s.readPayload()
	if err != nil {
		return 0, err
	}

	n := copy(b, data)

	if n != len(data) {
		s.readBuffer = data[n:]
	}

	return n, nil
}

func (s *Wrapper) Write(b []byte) (int, error) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	if err := s.writePayload(b); err != nil {
		return 0, err
	}

	return len(b), nil
}

func (s *Wrapper) readPayload() ([]byte, error) {
	lengthBytes := make([]byte, binary.MaxVarintLen64)
	if _, err := io.ReadFull(s.conn, lengthBytes); err != nil {
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
	ciphertext := make([]byte, length)
	if _, err := io.ReadFull(s.conn, ciphertext); err != nil {
		return nil, err
	}
	data, err := s.cipher.Decrypt(ciphertext)
	// todo: when server receives data and is unable to decrypt it, we should treat this
	// as an authentication failure and hang on the connection by reading infinitely
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Wrapper) writePayload(data []byte) error {
	ciphertext, err := s.cipher.Encrypt(data)
	if err != nil {
		return err
	}
	lengthBytes := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(lengthBytes, uint64(len(ciphertext)))
	if _, err := s.conn.Write(append(lengthBytes, ciphertext...)); err != nil {
		return err
	}
	return nil
}
