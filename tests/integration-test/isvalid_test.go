package integrationtest

import (
	"AuthDB/cmd/app/controller/helper"
	"AuthDB/cmd/app/repository"
	"AuthDB/tests/helpers"
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4"
)

// DB connection test (wrong connection)
func TestInitDBConn_InvalidURL(t *testing.T) {
	ctx := context.Background()
	invalidURL := "invalidUrl"

	_, err := repository.InitDBConn(ctx, invalidURL)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	expectedErrMsg := "failed to parse pg config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestIsValidPassword(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		user := &repository.User{
			Username: "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user: %v", err)
		}

		if !helper.IsValidPassword(user.Password) {
			t.Errorf("Password is not valid")
		}
		return nil
	})
}
