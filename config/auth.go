package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func generateAuthToken() string {
	return randString(32)
}

func randString(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return base64.RawStdEncoding.EncodeToString(buf)
}

func DecodeKeyString(k string) ([]byte, error) {
	key, err := base64.RawStdEncoding.DecodeString(k)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return key, fmt.Errorf("error: key is not 32 bytes")
	}
	return key, nil
}
