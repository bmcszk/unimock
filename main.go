// Unimock is a universal HTTP mock server for end-to-end testing.
// It can mock any HTTP-based service with flexible configuration.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bmcszk/unimock/pkg"
	"github.com/bmcszk/unimock/pkg/config"
)

const (
	// Signal channel buffer size
	signalChannelBuffer = 1
	// Shutdown timeout duration
	shutdownTimeout = 10 * time.Second
)

func main() {
	// Create a logger for main program logs
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration from environment variables
	serverConfig := config.FromEnv()

	logger.Info("starting unimock server",
		"port", serverConfig.Port,
		"config_path", serverConfig.ConfigPath,
		"log_level", serverConfig.LogLevel)

	// Load mock configuration from file
	uniConfig, err := config.LoadFromYAML(serverConfig.ConfigPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		panic(err)
	}

	// Start the server
	srv, err := pkg.NewServer(serverConfig, uniConfig)
	if err != nil {
		logger.Error("failed to initialize server", "error", err)
		panic(err)
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != context.Canceled {
			logger.Error("failed to start server", "error", err)
			panic(err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, signalChannelBuffer)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		panic(err)
	}

	logger.Info("server exited properly")
}
