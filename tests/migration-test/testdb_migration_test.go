package tests

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var (
	testDB *sql.DB
	dbURL  = "postgres://postgres:193566@localhost:5432/testdb?sslmode=disable"
)

func TestMigration(t *testing.T) {
	var err error

	testDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	migrationDir := filepath.Join(currentDir, "../../migrations")

	log.Printf("Current directory: %v", currentDir)
	log.Printf("Migration directory: %v", migrationDir)

	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		log.Fatalf("Migrations directory does not exist: %v", migrationDir)
	}

	err = goose.Up(testDB, migrationDir)
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	err = goose.Down(testDB, migrationDir)
	if err != nil {
		log.Fatalf("Failed to rollback migrations: %v", err)
	}

}
