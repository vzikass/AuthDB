package helper

import (
	"AuthDB/cmd/app/repository"
	"context"
	"fmt"
)

func GetUserByToken(token string) (u *repository.User, err error) {
	var rep repository.Repository
	userID, err := ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := rep.FindUserByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}
	return &user, nil
}
