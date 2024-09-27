package helpers

import (
	"AuthDB/cmd/app/repository"
	"context"
	"log"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	dbURL = "postgres://postgres:193566@localhost:5432/testdb?sslmode=disable"
)

// func to clear the db
func clearDatabase(t *testing.T, pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), "TRUNCATE users RESTART IDENTITY")
	if err != nil {
		t.Fatalf("Failed to clear database: %v", err)
	}
}

func RunWithTransactions(t *testing.T, fn func(tx pgx.Tx) error) {
	ctx := context.Background()
	
	pool, err := repository.InitDBConn(context.Background(), dbURL)
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
