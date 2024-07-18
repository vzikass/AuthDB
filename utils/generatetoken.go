package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomToken() string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(token)
}
