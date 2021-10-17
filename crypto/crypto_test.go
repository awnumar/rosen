package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lukechampine.com/frand"
)

func TestCipher(t *testing.T) {
	cipher, err := NewCipher(frand.Bytes(32))
	require.NoError(t, err)

	for i := 1; i <= 8; i++ {
		data := frand.Bytes(frand.Intn(4096))

		ciphertext, err := cipher.Encrypt(data)
		require.NoError(t, err)
		assert.Equal(t, len(data)+Overhead, len(ciphertext))

		plaintext, err := cipher.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(data, plaintext))
	}
}
