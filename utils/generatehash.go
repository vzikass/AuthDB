package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func GenerateHash(line string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(line), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}
