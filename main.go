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

	// Main app
	app := controller.NewApp(ctx, Dbpool)
	mainRouter := httprouter.New()
	app.Routes(mainRouter)

	// Start main app
	wg.Add(1)
	go func() {
		log.Println("Starting main HTTP server on port 4444")
		if err := http.ListenAndServe("0.0.0.0:4444", mainRouter); err != nil {
			log.Fatalf("Main server failed: %v", err)
		}
		defer wg.Done()
	}()
	wg.Wait()
}

// Не знаю как решить проблему с подлкючением к testdb, при сборке через докер тесты отрабатывают с кодом 0, сеть докера настроена правильно,
// строка dsn и вида url правильная, по логам миграции выполняются, могу подключиться к бд через контейнер с той же строкой. 
// транзакции выполняются