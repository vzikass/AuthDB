// verification of correct connection and execution of migrations
package testdbtest

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	testDB, err = sql.Open("postgres", "user=postgres dbname=testdb sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	currentDir, err := os.Getwd()
	if err != nil{
		log.Fatalf("Failed to get current directory: %v", err)
	}
	migrationDir := filepath.Join(currentDir, "./migrations")

	
	// up migrations 
	err = goose.Up(testDB, migrationDir)
	if err != nil{
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	code := m.Run()
	
	// down migrations 
	err = goose.Down(testDB, migrationDir)
	if err != nil{
		log.Fatalf("Failed to rollback migrations: %v", err)
	}
	
	os.Exit(code)
}
