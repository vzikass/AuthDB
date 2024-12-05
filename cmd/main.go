package main

import (
	"AuthDB/cmd/app/controller"
	"AuthDB/cmd/app/repository"
	"AuthDB/cmd/internal/kafka"
	useraccess "AuthDB/internal/api/user"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GoAdminGroup/go-admin/adapter/gorilla"
	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/examples/datamodel"
	"github.com/GoAdminGroup/go-admin/modules/config"
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/chartjs"
	_ "github.com/GoAdminGroup/themes/adminlte"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func initGoAdmin(router *mux.Router, dbURL string) (*engine.Engine, error) {
	// Parse DATABASE_URL for GoAdmin config
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %v", err)
	}

	user := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()
	host, port, _ := net.SplitHostPort(parsedURL.Host)
	database := strings.TrimPrefix(parsedURL.Path, "/")

	// Configure GoAdmin
	cfg := &config.Config{
		Env: config.EnvLocal,
		Databases: config.DatabaseList{
			"default": {
				Host:            host,
				Port:            port,
				User:            user,
				Pwd:             password,
				Name:            database,
				MaxIdleConns:    50,
				MaxOpenConns:    150,
				ConnMaxLifetime: time.Hour,
				Driver:          config.DriverPostgresql,
			},
		},
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		UrlPrefix: "admin",
		IndexUrl:  "/admin",
		Debug:     true,
		Theme:     "adminlte",
		Language:  language.EN,
	}

	// Initialize GoAdmin engine
	eng := engine.Default()
	eng.AddAdapter(&gorilla.Gorilla{})

	adminPlugin := admin.NewAdmin()

	if err := eng.AddConfig(cfg).
		AddGenerators().
		AddDisplayFilterXssJsFilter().
		AddGenerator("user", datamodel.GetUserTable).
		AddPlugins(adminPlugin).
		Use(router); err != nil {
		return nil, fmt.Errorf("failed to configure GoAdmin engine: %v", err)
	}

	template.AddComp(chartjs.NewChart())

	return eng, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load from .env file
	if err := godotenv.Load("/app/configs/db.env", "/app/configs/grpc.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	// Connect db
	dbpool, err := repository.InitDBConn(ctx, dbURL)
	if err != nil {
		log.Fatalf("Error initializing DB connection: %v\n", err)
	}
	defer dbpool.Close()

	// Init Kafka | producer | consumer
	producer, consumer, err := kafka.InitKafka()
	if err != nil {
		log.Fatalf("Error initializing Kafka: %v", err)
	}
	defer producer.Close()
	defer consumer.Close()

	// Main app
	// Initialize main application and router
	app := controller.NewApp(ctx, dbpool)
	mainRouter := mux.NewRouter()
	app.Routes(mainRouter)

	mainMux := http.NewServeMux()

	_, err = initGoAdmin(mainRouter, dbURL)
	if err != nil {
		log.Fatalf("Error initializing GoAdmin: %v", err)
	}
	// Main multiplexer with both admin and app routes
	mainMux.Handle("/", mainRouter) // Main app routes

	adminUser := &repository.User{
		Username: "admin",
		Password: "admin",
	}
	if err = adminUser.AddAdminUser(context.Background(), dbpool, nil); err != nil {
		log.Fatalf("Failed to add admin user: %v", err)
	}

	// HTTP Server
	server := &http.Server{
		Addr:    "0.0.0.0:4444",
		Handler: mainMux,
	}

	go func() {
		log.Println("Starting server on port 4444")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// gRPC Server
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatalf("GRPC_PORT not set")
	}
	// Create an AccessService instance
	accessService := &useraccess.AccessService{}
	if err := useraccess.StartGRPCServer(":"+port, accessService); err != nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}

	// Graceful shutdown
	// we need to reserve to buffer size 1, so the notifier are not blocked
	exit := make(chan os.Signal, 1)
	// The operating system sends a shutdown signal to a process when it wants to terminate it gracefully
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	// program is blocked until the channel receives a signal.
	// waiting to receive a signal such as os.Interrupt or syscall.SIGTERM
	go func() {
		<-exit
		log.Println("Shutting down servers...")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// http server shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP Server Shutdown Failed: %v", err)
		}

		// Kafka shutdown
		log.Println("Closing Kafka producer and consumer...")
		if err := producer.Close(); err != nil {
			log.Printf("Kafka producer close failed: %v", err)
		}
		if err := consumer.Close(); err != nil {
			log.Printf("Kafka consumer close failed: %v", err)
		}

		// Close the db connection
		log.Println("Closing database connection pool...")
		dbpool.Close()

		// cancel app ctx
		cancel()

		log.Println("Shutdown complete.")
	}()
}
