package helper

import (
	"AuthDB/cmd/app/repository"
	"context"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

func parseToken(token string) (int, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || !parsedToken.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return 0, fmt.Errorf("user ID not found in token")
	}

	userID := int(claims["user_id"].(float64))
	return userID, nil
}

func GetUserByToken(token string) (u *repository.User, err error) {
	var rep repository.Repository
	userID, err := parseToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := rep.FindUserByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}
	return &user, nil
}