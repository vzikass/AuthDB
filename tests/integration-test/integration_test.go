// Integration tests to check the connection to the test database
// Work with transactions
package integrationtest

import (
	"AuthDB/cmd/app/controller/helper"
	"AuthDB/cmd/app/repository"
	"AuthDB/tests/helpers"
	"AuthDB/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4"
)

var (
	ErrUserNotFound = errors.New("user not found")
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

// -----------------------

// User tests (using test db)
func TestNewUser(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		login := "testuser"
		password := "qwerty123"
		email := "testuser@example.com"
		user, err := repository.NewUser(login, email, password)
		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}

		ctx := context.Background()
		if err := user.Add(ctx, tx); err != nil {
			return fmt.Errorf("failed to add user: %v", err)
		}

		if user.Login != login || user.Email != email {
			t.Errorf("User data is incorrect")
		}
		if !utils.CompareHashPassword(password, user.Password) {
			t.Errorf("Password hash does not match")
		}
		return nil
	})
}

func TestGetAllUsers(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		user := &repository.User{
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user for testing: %v", err)
		}

		users, err := repository.GetAllUsers(ctx, tx)
		if err != nil {
			t.Fatalf("Failed to get all users: %v", err)
		}
		if len(users) == 0 {
			t.Errorf("No users found")
		}
		return err
	})
}

func TestGetUserByID(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}
		user := &repository.User{
			ID:       1,
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user for testing: %v", err)
		}
		t.Logf("User added with ID: %d", user.ID)
		userID := user.ID
		u, err := repo.GetByID(ctx, tx, userID)
		if err != nil {
			log.Fatalf("Failed to get user by id: %v", err)
		}
		if u.ID != userID {
			t.Errorf("Expected user ID %d, got %d", userID, user.ID)
		}
		return err
	})
}

func TestAddUser(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}
		user := &repository.User{
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user: %v", err)
		}

		exist, err := repo.UserExist(ctx, tx, user.Login, user.Email)
		if err != nil {
			t.Errorf("User not found")
		}

		if !exist {
			t.Errorf("User was not found after addition")
		} else {
			t.Log("User successfully added!")
		}
		return err
	})
}

func TestUpdateUser(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}
		user := &repository.User{
			ID:       1,
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user: %v", err)
		}

		user.Login = "updateduser"
		user.Password = "newpassword123"
		user.Email = "updated@example.com"
		if err := user.UpdateByID(ctx, tx); err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		updatedUser, err := repo.GetByID(ctx, tx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}
		if updatedUser.Login != "updateduser" || updatedUser.Email != "updated@example.com" {
			t.Errorf("User data was not updated correctly. Got %+v", updatedUser)
		}
		if utils.CompareHashPassword("newpassword123", updatedUser.Password) {
			t.Errorf("Password was not updated correctly")
		} else {
			t.Log("User successfully updated!")
		}
		return err
	})
}

func TestDeleteUser(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}
		user := &repository.User{
			ID:       1,
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user for testing: %v", err)
		}

		if err := user.DeleteByID(ctx, tx, user.ID); err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		u, err := repo.GetByID(ctx, tx, user.ID)
		if err == nil {
			t.Fatalf("User was not deleted. User found with ID: %d", u.ID)
		}
		if errors.Is(err, ErrUserNotFound) {
			t.Errorf("Unexpected error when getting user by ID: %v", err)
		} else {
			t.Log("User successfully deleted!")
		}
		return err
	})
}

func TestLogin(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}
		login := "testuser"
		user := &repository.User{
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user: %v", err)
		}

		u, err := repo.Login(ctx, tx, user.Login)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if u.Login != login {
			t.Errorf("Expected login %s, got %s", login, user.Login)
		}
		return err
	})
}

func TestUserExist(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}
		user := &repository.User{
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user: %v", err)
		}

		exist, err := repo.UserExist(ctx, tx, user.Login, user.Email)
		if err != nil {
			t.Fatalf("Failed to check exist user: %v", err)
		}

		if !exist {
			t.Errorf("User should exist, but was not found")
		} else {
			t.Log("User found!")
		}
		return err
	})
}

func TestUserNotExist(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		repo := &repository.Repository{}

		login := "nonexistentuser"
		email := "nonexistent@example.com"

		exist, err := repo.UserExist(ctx, tx, login, email)
		if err != nil {
			t.Fatalf("Failed to check if user exists: %v", err)
		}
		if exist {
			t.Errorf("User should not exist, but was found")
		} else {
			t.Log("User not found!")
		}
		return err
	})
}

func TestIsValidPassword(t *testing.T) {
	helpers.RunWithTransactions(t, func(tx pgx.Tx) error {
		ctx := context.Background()
		user := &repository.User{
			Login:    "testuser",
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
