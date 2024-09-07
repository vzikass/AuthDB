package utils

import "golang.org/x/crypto/bcrypt"

// Compare hashedPassword and just password
func CompareHashPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}