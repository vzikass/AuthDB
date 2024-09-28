package helpers

import (
	"AuthDB/cmd/app/repository"
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pressly/goose/v3"
)

var (
	DBURL = "postgres://postgres:193566@localhost:5433/testdb?sslmode=disable"
)

func UpMigrations(t *testing.T) error{
	testDB, err := sql.Open("postgres", DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	migrationDir := filepath.Join(currentDir, "../../migrations")

	err = goose.Up(testDB, migrationDir)
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	return nil
}

// func to clear the db
func clearDatabase(t *testing.T, pool *pgxpool.Pool) {
	if err := UpMigrations(t); err != nil{
		t.Fatalf("Failed to up migrations: %v", err)
	}
	_, err := pool.Exec(context.Background(), "TRUNCATE users RESTART IDENTITY")
	if err != nil {
		t.Fatalf("Failed to clear database: %v", err)
	}
}

func RunWithTransactions(t *testing.T, fn func(tx pgx.Tx) error) {
	ctx := context.Background()

	pool, err := repository.InitDBConn(context.Background(), DBURL)
	if err != nil {
		log.Fatalf("Error initializing Test DB connection: %v\n", err)
	}
	defer pool.Close()

	// Clear DB before starting transaction
	clearDatabase(t, pool)

	// Start transaction
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	err = fn(tx)
}
