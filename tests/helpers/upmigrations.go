package helpers

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

func UpMigrations(t *testing.T) error {
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
