package crypto

import (
	"bytes"
	"testing"

	"github.com/matryer/is"
	"lukechampine.com/frand"
)

func TestCipher(t *testing.T) {
	is := is.New(t)

	cipher, err := NewCipher(frand.Bytes(32))
	is.NoErr(err)

	for i := 1; i <= 8; i++ {
		data := frand.Bytes(frand.Intn(4096))

		ciphertext, err := cipher.Encrypt(data)
		is.NoErr(err)
		is.Equal(len(data)+Overhead, len(ciphertext))

		plaintext, err := cipher.Decrypt(ciphertext)
		is.NoErr(err)
		is.True(bytes.Equal(data, plaintext))
	}
}
