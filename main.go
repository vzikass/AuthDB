package main

import (
	"context"
	"exercise/app/controller"
	"exercise/app/repository"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	ctx := context.Background()

	// Connect to database (Postgres)
	dbpool, err := repository.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("Error init DB connection: %v\n", err)
	}
	defer dbpool.Close()

	// Routes
	a := controller.NewApp(ctx, dbpool)
	r := httprouter.New()
	a.Routes(r)

	// Start Local Server
	err = http.ListenAndServe("localhost:4444", r)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
