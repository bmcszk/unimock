package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/bmcszk/unimock/pkg"
	"github.com/bmcszk/unimock/pkg/config"
)

// This example demonstrates how to use Unimock as a library with in-code configuration
func main() {
	// Set up a logger for the main program
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Unimock with in-code configuration")

	// Step 1: Create a mock configuration with API endpoints
	mockConfig := &config.MockConfig{
		Sections: map[string]config.Section{
			"api": {
				// Match paths like /api/123
				PathPattern: "/api/*",
				// Extract IDs from these paths in JSON body
				BodyIDPaths: []string{"/id", "/data/id"},
				// Extract ID from this header if present
				HeaderIDName: "X-API-ID",
			},
			"users": {
				// Match paths like /users/456
				PathPattern: "/users/*",
				// Extract IDs from these paths in JSON body
				BodyIDPaths: []string{"/id", "/userId", "/user/id"},
				// Extract ID from this header if present
				HeaderIDName: "X-User-ID",
			},
		},
	}

	// Step 2: Create server configuration
	serverConfig := &config.ServerConfig{
		Port:       "8081",        // HTTP port to listen on
		LogLevel:   "debug",       // Log level (debug, info, warn, error)
		ConfigPath: "config.yaml", // Not used for in-code config, but required
	}

	// Step 3: Initialize the server
	logger.Info("Creating server with configuration")
	srv, err := pkg.NewServer(serverConfig, mockConfig)

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Step 4: Start the server
	logger.Info("Server created successfully", "port", serverConfig.Port)

	go func() {
		logger.Info("Starting HTTP server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				logger.Error("Server error", "error", err)
			}
		}
	}()

	// Step 5: Wait for shutdown signal
	const signalChannelBuffer = 1
	quit := make(chan os.Signal, signalChannelBuffer)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server")

	// Step 6: Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
