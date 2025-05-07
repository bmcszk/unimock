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
	mockConfig, err := config.LoadFromYAML(serverConfig.ConfigPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		panic(err)
	}

	// Start the server
	srv, err := pkg.NewServer(serverConfig, mockConfig)
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
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		panic(err)
	}

	logger.Info("server exited properly")
}
