package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)


const maxPasswordLength = 72

func GenerateToken(password string) (string, error) {
	if len(password) > maxPasswordLength {
		return "", fmt.Errorf("password length exceeds %d bytes", maxPasswordLength)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate hash token: %w", err)
	}
	return string(hashedPassword), nil
}
