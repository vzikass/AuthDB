package utils

import (
	"golang.org/x/crypto/bcrypt"
)
// Generating hash (for password, tokens)
func GenerateHash(line string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(line), bcrypt.DefaultCost) // DefaultCost = 10
	if err != nil {
		return "", err
	}
	return string(hash), err
}
