package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/server"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nuklai/nuklai-feed/config"
	"github.com/nuklai/nuklai-feed/manager"
	frpc "github.com/nuklai/nuklai-feed/rpc"
	"go.uber.org/zap"
)

var (
	httpConfig = server.HTTPConfig{
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
)

func fatal(l logging.Logger, msg string, fields ...zap.Field) {
	l.Fatal(msg, fields...)
	os.Exit(1)
}

// HealthHandler responds with a simple health check status
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	err := godotenv.Overload() // Overload the environment variables with those from the .env file
	if err != nil {
		utils.Outf("{{red}}Error loading .env file{{/}}: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Loaded environment variables from .env file")

	logFactory := logging.NewFactory(logging.Config{
		DisplayLevel: logging.Info,
	})
	l, err := logFactory.Make("main")
	if err != nil {
		utils.Outf("{{red}}unable to initialize logger{{/}}: %v\n", err)
		os.Exit(1)
	}
	log := l
	log.Info("Logger initialized")

	// Load config from environment variables
	config, err := config.LoadConfigFromEnv()
	if err != nil {
		fatal(log, "cannot load config from environment variables", zap.Error(err))
	}
	log.Info("Config loaded from environment variables")

	// Load recipient
	if _, err := config.RecipientAddress(); err != nil {
		fatal(log, "cannot parse recipient address", zap.Error(err))
	}
	log.Info("Loaded feed recipient", zap.String("address", config.Recipient))

	// Create server
	listenAddress := net.JoinHostPort(config.HTTPHost, fmt.Sprintf("%d", config.HTTPPort))
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fatal(log, "cannot create listener", zap.Error(err))
	}
	log.Info("Created listener", zap.String("address", listenAddress))

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         listenAddress,
		Handler:      mux,
		ReadTimeout:  httpConfig.ReadTimeout,
		WriteTimeout: httpConfig.WriteTimeout,
		IdleTimeout:  httpConfig.IdleTimeout,
	}

	// Add health check handler
	mux.HandleFunc("/health", HealthHandler)
	log.Info("Health handler added")

	// Retry mechanism for PostgreSQL connection
	var db *sql.DB
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresDBName, config.PostgresSSLMode))
		if err != nil {
			log.Warn("Error opening database", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		err = db.Ping()
		if err == nil {
			break
		}
		log.Warn("Database not ready, retrying...", zap.Error(err))
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		fatal(log, "could not connect to the database", zap.Error(err))
	}
	log.Info("Database connection established")

	// Start manager with context handling
	manager, err := manager.New(log, config, db)
	if err != nil {
		fatal(log, "cannot create manager", zap.Error(err))
	}
	log.Info("Manager created")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		log.Info("Starting manager")
		if err := manager.Run(ctx); err != nil {
			log.Error("Manager error", zap.Error(err))
		}
	}()

	// Add feed handler
	feedServer := frpc.NewJSONRPCServer(manager)
	handler, err := server.NewHandler(feedServer, "feed")
	if err != nil {
		fatal(log, "cannot create handler", zap.Error(err))
	}
	mux.Handle("/", handler)
	log.Info("Feed handler added")

	// Start server
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info("Triggering server shutdown", zap.Any("signal", sig))
		cancel() // Ensure context cancellation cascades down
		_ = srv.Shutdown(ctx)
	}()
	log.Info("Server starting")

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed", zap.Error(err))
	}
	log.Info("Server exited")
}
