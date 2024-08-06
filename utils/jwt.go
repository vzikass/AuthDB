package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(1 * time.Hour).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func ParseJWT(tokenString string) (*jwt.Token, *jwt.MapClaims, error){
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil{
		return nil, nil, err
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok{
		return nil, nil, err
	}
	return token, claims, nil
}
