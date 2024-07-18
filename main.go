package main

import (
	"context"
	"encoding/json"
	"exercise/app/controller"
	"exercise/app/repository"
	"io"
	"log"
	"net/http"
	"os"

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

	// Load Fixtures
	users, err := loadfixtures("fixtures/fixtures.json")
	if err != nil{
		log.Fatalf("Unable to load fixtures: %v\n", err)
	}
	for _, user := range users{
		_, err = repository.AddFixtures(ctx)
		if err != nil{
			log.Fatalf("Failed to insert user %s: %v", user.Login, err)
		}
	}

	// Start Local Server
	err = http.ListenAndServe("localhost:4444", r)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func loadfixtures(filename string) ([]repository.User, error){
	var users []repository.User
	file, err := os.Open(filename)
	if err != nil{
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil{
		return nil, err
	}
	err = json.Unmarshal(data, &users)
	if err != nil{
		return nil, err
	}
	return users, nil
}