// Simple tests of all the functionality of my project using transactions
package repository

import (
	"AuthDB/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrUserNotFound = errors.New("User not found")
	DbURL           = "postgres://postgres:193566@testdb:5432/testdb?sslmode=disable"
)

func clearDatabase(t *testing.T, pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), "TRUNCATE users RESTART IDENTITY")
	if err != nil {
		t.Fatalf("Failed to clear database: %v", err)
	}
}

func waitForDB(t *testing.T, dsn string, maxAttempts int, delay time.Duration){
	for attempts := 1; attempts <= maxAttempts; attempts++{
		pool, err := InitDBConn(context.Background(), dsn)
		if err == nil{
			pool.Close()
			return
		}
		time.Sleep(delay)
	}
	t.Fatalf("Database is not available after %d attempts", maxAttempts)
}

func RunWithTransactions(t *testing.T, fn func(tx pgx.Tx) error) {
	os.Setenv("TESTDB_URL", "postgres://postgres:193566@testdb:5432/testdb?sslmode=disable")
	pool, err := InitDBConn(context.Background(), os.Getenv("TESTDB_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()
	
	// Clear DB before starting transaction
	clearDatabase(t, pool)

	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			panic(r)
		} else if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	err = fn(tx)
}

// Database tests
func TestInitDBConn_InvalidURL(t *testing.T) {
	ctx := context.Background()
	invalidURL := "invalidUrl"

	_, err := InitDBConn(ctx, invalidURL)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	expectedErrMsg := "failed to parse pg config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
func TestInitDBConn_UnreachableServer(t *testing.T) {
	ctx := context.Background()
	testdbURL := "postgres:postgres:193566@testdb:9753/testdb?sslmode=disable"

	_, err := InitDBConn(ctx, testdbURL)
	if err == nil {
		t.Fatalf("Expected error but got nil")
	}

	expectedErrMsg := "failed to connect config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestInitDBConn_InsufficientPrivileges(t *testing.T) {
	ctx := context.Background()
	testdbURL := "postgres:postgres:WRONGPASSWORD@testdb:5432/testdb?sslmode=disable"

	_, err := InitDBConn(ctx, testdbURL)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	expectedErrMsg := "failed to connect config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

// -----------------------

// User tests
func TestNewUser(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) error {
		login := "testuser"
		password := "qwerty123"
		email := "testuser@example.com"
		user, err := NewUser(login, email, password)
		if err != nil {
			return fmt.Errorf("Failed to create user: %v", err)
		}

		ctx := context.Background()
		if err := user.Add(ctx, tx); err != nil {
			return fmt.Errorf("Failed to add user: %v", err)
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
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		user := &User{
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user for testing: %v", err)
		}

		users, err := GetAllUsers(ctx, tx)
		if err != nil {
			t.Fatalf("Failed to get all users: %v", err)
		}
		if len(users) == 0 {
			t.Errorf("No users found")
		}
		return
	})
}

func TestGetUserByID(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}
		user := &User{
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
		return
	})
}

func TestAddUser(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}
		user := &User{
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
		return
	})
}

func TestUpdateUser(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}
		user := &User{
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
		if err := user.Update(ctx, tx); err != nil {
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
		return
	})
}

func TestDeleteUser(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}
		user := &User{
			ID:       1,
			Login:    "testuser",
			Password: "qwerty123",
			Email:    "testuser@example.com",
		}

		if err := user.Add(ctx, tx); err != nil {
			t.Fatalf("Failed to add user for testing: %v", err)
		}

		if err := user.Delete(ctx, tx, user.ID); err != nil {
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
		return
	})
}

func TestLogin(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}
		login := "testuser"
		user := &User{
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
		return
	})
}

func TestUserExist(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}
		user := &User{
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
		return
	})
}

func TestUserNotExist(t *testing.T) {
	waitForDB(t, os.Getenv("TESTDB_URL"), 10, 2*time.Second)
	RunWithTransactions(t, func(tx pgx.Tx) (err error) {
		ctx := context.Background()
		repo := &Repository{pool: TestDbpool}

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
		return
	})
}

// --------------------

// Utils tests

func TestCompareHashPassword(t *testing.T) {
	password := "secretpassword"

	hashedPassword, err := utils.GenerateHash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if !utils.CompareHashPassword(password, hashedPassword) {
		t.Errorf("CompareHashPassword failed for valid password")
	}

	wrongPassword := "wrongpassword"
	if utils.CompareHashPassword(wrongPassword, hashedPassword) {
		t.Errorf("CompareHashPassword succeeded for invalid password")
	}
}

// Test jwt token
func TestGenerateJWT(t *testing.T) {
	userID := "testUser"

	token, err := utils.GenerateJWT(userID)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if len(token) == 0 {
		t.Errorf("Generated JWT is empty")
	}
}

func TestParseJWT(t *testing.T) {
	userID := "testUser"
	tokenString, err := utils.GenerateJWT(userID)

	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	token, claims, err := utils.ParseJWT(tokenString)
	if err != nil {
		t.Fatalf("Failed to ParseJWT: %v", err)
	}

	if !token.Valid {
		t.Errorf("Token is not valid: %v", err)
	}

	if (*claims)["userID"] != userID {
		t.Errorf("Expected userID %v, got %v", userID, (*claims)["userID"])
	}

	exp, ok := (*claims)["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		t.Errorf("JWT token has expired")
	}
}
