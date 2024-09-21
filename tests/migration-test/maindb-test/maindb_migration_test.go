// verification of correct connection and execution of migrations
package tests

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var mainDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	mainDB, err = sql.Open("postgres", "user=postgres dbname=AuthDB sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer mainDB.Close()

	// up migrations
	err = goose.Up(mainDB, "/migrations")
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	code := m.Run()

	// down migrations
	err = goose.Down(mainDB, "/migrations")
	if err != nil {
		log.Fatalf("Failed to rollback migrations: %v", err)
	}
	os.Exit(code)
}
