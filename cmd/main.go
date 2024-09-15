package main

import (
	"AuthDB/cmd/app/controller"
	"AuthDB/cmd/app/repository"
	"AuthDB/cmd/internal/kafka"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load from .env file
	if err := godotenv.Load("./db.env"); err != nil {
		log.Fatalf("Failed to load db.env: %v", err)
	}

	// Connect db
	dbpool, err := repository.InitDBConn(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error initializing DB connection: %v\n", err)
	}

	// Init Kafka | producer | consumer
	kafka.InitKafka()
	defer kafka.Producer.Close()
	defer kafka.Consumer.Close()

	// Main app
	app := controller.NewApp(ctx, dbpool)
	mainRouter := httprouter.New()
	app.Routes(mainRouter)

	server := &http.Server{
		Addr: "0.0.0.0:4444",
		Handler: mainRouter,
	}

	// Start main app
	go func() {
		log.Println("Starting main HTTP server on port 4444")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Main server failed: %v", err)
		}
	}()
	
	// --------------
	// Graceful shutdown? :)

	// we need to reserve to buffer size 1, so the notifier are not blocked
	exit := make(chan os.Signal, 1)
	// The operating system sends a shutdown signal to a process when it wants to terminate it gracefully
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	// program is blocked until the channel receives a signal. 
	// waiting to receive a signal such as os.Interrupt or syscall.SIGTERM
	<- exit
}
