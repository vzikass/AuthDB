package main

import (
	"context"
	"AuthDB/app/controller"
	"AuthDB/app/repository"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

func main() {
	ctx := context.Background()

	// Load from .env file
	if err := godotenv.Load("db.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	dbURL, exist := os.LookupEnv("DATABASE_URL")
	if !exist {
		log.Print("No variables with the same name were found")
	}
	// Connect db (postgres)
	dbpool, err := repository.InitDBConn(ctx, dbURL)
	if err != nil {
		log.Fatalf("Error init DB connection: %v\n", err)
	}
	defer dbpool.Close()

	// Routes
	a := controller.NewApp(ctx, dbpool)
	r := httprouter.New()
	a.Routes(r)

	// Start Local Server
	err = http.ListenAndServe("0.0.0.0:4444", r)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
