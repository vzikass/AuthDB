package main

import (
	"AuthDB/cmd/app/controller"
	"AuthDB/cmd/app/repository"
	"AuthDB/cmd/internal/config"
	"AuthDB/cmd/internal/kafka"
	useraccess "AuthDB/internal/api/user"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load from .env file
	if err := config.Load("/app/configs/db.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Connect db
	dbpool, err := repository.InitDBConn(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error initializing DB connection: %v\n", err)
	}

	// Init Kafka | producer | consumer
	producer, consumer := kafka.InitKafka()
	defer producer.Close()
	defer consumer.Close()

	// Main app
	app := controller.NewApp(ctx, dbpool)
	mainRouter := httprouter.New()
	app.Routes(mainRouter)

	server := &http.Server{
		Addr:    "0.0.0.0:4444",
		Handler: mainRouter,
	}

	// Start main app
	go func() {
		log.Println("Starting main HTTP server on port 4444")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Main server failed: %v", err)
		}
	}()

	// Create an AccessService instance
	accessService := &useraccess.AccessService{}

	// Load env from .env
	if err := config.Load("/app/configs/grpc.env"); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatalf("GRPC_PORT not set")
	}

	// Running grpc in a separate goroutine
	go func() {
		if err := useraccess.StartGRPCServer(":"+port, accessService); err != nil {
			log.Fatalf("Failed to start grpc server: %v", err)
		}
	}()

	// Graceful shutdown
	// we need to reserve to buffer size 1, so the notifier are not blocked
	exit := make(chan os.Signal, 1)
	// The operating system sends a shutdown signal to a process when it wants to terminate it gracefully
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	// program is blocked until the channel receives a signal.
	// waiting to receive a signal such as os.Interrupt or syscall.SIGTERM
	<-exit
}
