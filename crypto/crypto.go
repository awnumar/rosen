package crypto

import (
	"crypto/rand"

	"golang.org/x/crypto/chacha20poly1305"
)

const Overhead = chacha20poly1305.NonceSizeX + chacha20poly1305.Overhead

func Encrypt(plaintext, key []byte) ([]byte, error) {
	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, cipher.NonceSize(), cipher.NonceSize()+len(plaintext)+cipher.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return cipher.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext, key []byte) ([]byte, error) {
	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < cipher.NonceSize() {
		return nil, err
	}

	return cipher.Open(nil, ciphertext[:cipher.NonceSize()], ciphertext[cipher.NonceSize():], nil)
}
