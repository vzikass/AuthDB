package tests

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/pressly/goose/v3"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	testDB, err = sql.Open("postgres", "user=youruser dbname=testdb sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	err = goose.Up(testDB, "/migrations")
	if err != nil{
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	code := m.Run()

	err = goose.Down(testDB, "/migrations")
	if err != nil{
		log.Fatalf("Failed to rollback migrations: %v", err)
	}
	
	os.Exit(code)
}
