// verification of correct connection and execution of migrations
package maindbtest

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var mainDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	mainDB, err = sql.Open("postgres", "user=postgres dbname=Authdb sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer mainDB.Close()

	currentDir, err := os.Getwd()
	if err != nil{
		log.Fatalf("Failed to get current directory: %v", err)
	}
	migrationDir := filepath.Join(currentDir, "./migations")

	// up migrations
	err = goose.Up(mainDB, migrationDir)
	if err != nil{
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	code := m.Run()

	// down migrations
	err = goose.Down(mainDB, migrationDir)
	if err != nil{
		log.Fatalf("Failed to rollback migrations: %v", err)
	}
	os.Exit(code)
}
