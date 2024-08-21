package main

import (
	"AuthDB/cmd/app/controller"
	"AuthDB/cmd/app/repository"
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

var wg sync.WaitGroup

func main() {
	ctx := context.Background()
	// Load from .env file
	if err := godotenv.Load("db.env"); err != nil{
		log.Fatalf("Failed to load db.env: %v", err)
	}

	// Connect main db 
	Dbpool, err := repository.InitDBConn(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error initializing DB connection: %v\n", err)
	}

	// Connect test DB
	TestDbpool, err := repository.InitDBConn(ctx, os.Getenv("TESTDB_URL"))
	if err != nil {
		log.Fatalf("Error initializing Test DB connection: %v\n", err)
	}

	// Main app
	app := controller.NewApp(ctx, Dbpool)
	mainRouter := httprouter.New()
	app.Routes(mainRouter)

	// Test app
	testApp := controller.NewApp(ctx, TestDbpool)
	testRouter := httprouter.New()
	testApp.Routes(testRouter)

	// Migratoins
	// runMigrations(os.Getenv("TESTDB_URL"))

	// Start main app
	wg.Add(1)
	go func() {
		log.Println("Starting main HTTP server on port 4444")
		if err := http.ListenAndServe("0.0.0.0:4444", mainRouter); err != nil {
			log.Fatalf("Main server failed: %v", err)
		}
		defer wg.Done()
	}()
	// Start test server
	wg.Add(1)
	go func() {
		log.Println("Starting test HTTP server on port 4445")
		err := http.ListenAndServe("0.0.0.0:4445", testRouter)
		if err != nil {
			log.Fatalf("Test server failed: %v", err)
		}
		log.Println("Test server stopped")
		defer wg.Done()
	}()

	wg.Wait()
}
