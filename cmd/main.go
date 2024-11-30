package main

import (
	"AuthDB/cmd/app/controller"
	"AuthDB/cmd/app/repository"
	"AuthDB/cmd/internal/kafka"
	useraccess "AuthDB/internal/api/user"
	"context"
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load from .env file
	if err := godotenv.Load("/app/configs/db.env", "/app/configs/grpc.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Connect db
	dbpool, err := repository.InitDBConn(ctx, os.Getenv("DATABASE_URL"))
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

	// Parse DATABASE_URL for GoAdmin config
	parsedURL, err := url.Parse(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to parse DATABASE_URL: %v", err)
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
	cfg.Debug = true
	// Set up GoAdmin with plugins and templates
	eng := engine.Default()
	eng.AddAdapter(&gorilla.Gorilla{})

	adminPlugin := admin.NewAdmin()

	if err := eng.AddConfig(cfg).
		AddGenerators().
		AddDisplayFilterXssJsFilter().
		AddGenerator("user", datamodel.GetUserTable).
		AddPlugins(adminPlugin).
		Use(mainRouter); err != nil {
		panic(err)
	}
	template.AddComp(chartjs.NewChart())

	// Main multiplexer with both admin and app routes
	mainMux := http.NewServeMux()
	mainMux.Handle("/", mainRouter) // Main app routes
	mainMux.Handle("/admin/", http.StripPrefix("/admin", controller.AdminMux))

	adminUser := &repository.User{
		Username: "admin",
		Password: "admin",
	}
	err = adminUser.AddAdminUser(context.Background(), dbpool, nil)
	if err != nil {
		log.Fatal("Failed to add admin user:", err)
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
	go func() {
		port := os.Getenv("GRPC_PORT")
		if port == "" {
			log.Fatalf("GRPC_PORT not set")
		}
		// Create an AccessService instance
		accessService := &useraccess.AccessService{}
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
