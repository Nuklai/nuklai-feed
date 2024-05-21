package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/server"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/joho/godotenv"
	"github.com/nuklai/nuklai-feed/config"
	"github.com/nuklai/nuklai-feed/manager"
	frpc "github.com/nuklai/nuklai-feed/rpc"
	"go.uber.org/zap"
)

var (
	allowedOrigins  = []string{"*"}
	allowedHosts    = []string{"*"}
	shutdownTimeout = 30 * time.Second
	httpConfig      = server.HTTPConfig{
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

func main() {
	err := godotenv.Load()
	if err != nil {
		utils.Outf("{{red}}Error loading .env file{{/}}: %v\n", err)
		os.Exit(1)
	}

	logFactory := logging.NewFactory(logging.Config{
		DisplayLevel: logging.Info,
	})
	l, err := logFactory.Make("main")
	if err != nil {
		utils.Outf("{{red}}unable to initialize logger{{/}}: %v\n", err)
		os.Exit(1)
	}
	log := l

	// Load config from environment variables
	config, err := config.LoadConfigFromEnv()
	if err != nil {
		fatal(log, "cannot load config from environment variables", zap.Error(err))
	}

	// Load recipient
	if _, err := config.RecipientAddress(); err != nil {
		fatal(log, "cannot parse recipient address", zap.Error(err))
	}
	log.Info("loaded feed recipient", zap.String("address", config.Recipient))

	// Create server
	listenAddress := net.JoinHostPort(config.HTTPHost, fmt.Sprintf("%d", config.HTTPPort))
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fatal(log, "cannot create listener", zap.Error(err))
	}
	srv, err := server.New("", log, listener, httpConfig, allowedOrigins, allowedHosts, shutdownTimeout)
	if err != nil {
		fatal(log, "cannot create server", zap.Error(err))
	}

	// Start manager with context handling
	manager, err := manager.New(log, config)
	if err != nil {
		fatal(log, "cannot create manager", zap.Error(err))
	}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := manager.Run(ctx); err != nil {
			log.Error("manager error", zap.Error(err))
		}
	}()

	// Add feed handler
	feedServer := frpc.NewJSONRPCServer(manager)
	handler, err := server.NewHandler(feedServer, "feed")
	if err != nil {
		fatal(log, "cannot create handler", zap.Error(err))
	}
	if err := srv.AddRoute(handler, "feed", ""); err != nil {
		fatal(log, "cannot add feed route", zap.Error(err))
	}

	// Start server
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info("triggering server shutdown", zap.Any("signal", sig))
		cancel() // Ensure context cancellation cascades down
		_ = srv.Shutdown()
	}()
	log.Info("server exited", zap.Error(srv.Dispatch()))
}
