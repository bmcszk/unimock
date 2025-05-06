package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	port     string
	idPaths  []string
	idHeader string
	logLevel string
}

func parseConfig() *config {
	cfg := &config{}
	flag.StringVar(&cfg.port, "port", "8080", "Port to listen on")
	flag.StringVar(&cfg.idHeader, "id-header", "X-Resource-ID", "Header name to look for ID")
	flag.StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Default ID paths
	cfg.idPaths = []string{"//id", "//@id"}

	return cfg
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func main() {
	// Parse configuration
	cfg := parseConfig()

	// Setup logger
	logger := setupLogger(cfg.logLevel)
	logger.Info("starting server",
		"port", cfg.port,
		"id_header", cfg.idHeader,
		"id_paths", cfg.idPaths,
		"log_level", cfg.logLevel)

	// Create a new storage
	storage := NewStorage()

	// Create a new handler
	handler := NewHandler(storage, cfg.idPaths, cfg.idHeader, logger)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server",
				"error", err)
			os.Exit(1)
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
		logger.Error("server forced to shutdown",
			"error", err)
		os.Exit(1)
	}

	logger.Info("server exited properly")
}
