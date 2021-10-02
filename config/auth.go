package config

import (
	"crypto/rand"
	"encoding/base64"
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
