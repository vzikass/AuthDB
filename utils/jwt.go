package utils

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	SecretKey = []byte("secret")
)

func GenerateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	// map to store claims
	claims := token.Claims.(jwt.MapClaims)
	// set token
	claims["username"] = username
	// expiration time
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
		return "", nil
	}
	return tokenString, nil
}