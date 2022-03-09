package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const Overhead = chacha20poly1305.Overhead + chacha20poly1305.NonceSizeX

type Cipher struct {
	cipher.AEAD
}

func NewCipher(key []byte) (*Cipher, error) {
	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	return &Cipher{cipher}, nil
}

func (c *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.NonceSize(), len(plaintext)+Overhead)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return c.Seal(nonce, nonce, plaintext, nil), nil
}

func (c *Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < c.NonceSize() {
		return nil, fmt.Errorf("error: ciphertext too short")
	}
	return c.Open(nil, ciphertext[:c.NonceSize()], ciphertext[c.NonceSize():], nil)
}
