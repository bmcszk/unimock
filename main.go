package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bmcszk/unimock/config"
)

type Envs struct {
	port       string
	configPath string
	logLevel   string
}

func parseConfig() *Envs {
	envs := &Envs{}

	// Get port from environment variable, default to 8080
	if port := os.Getenv("UNIMOCK_PORT"); port != "" {
		envs.port = port
	} else {
		envs.port = "8080"
	}

	// Get config path from environment variable, default to config.yaml
	if configPath := os.Getenv("UNIMOCK_CONFIG"); configPath != "" {
		envs.configPath = configPath
	} else {
		envs.configPath = "config.yaml"
	}

	// Get log level from environment variable, default to info
	if logLevel := os.Getenv("UNIMOCK_LOG_LEVEL"); logLevel != "" {
		envs.logLevel = logLevel
	} else {
		envs.logLevel = "info"
	}

	return envs
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
	conf := parseConfig()

	// Setup logger
	logger := setupLogger(conf.logLevel)
	logger.Info("starting server",
		"port", conf.port,
		"config_path", conf.configPath,
		"log_level", conf.logLevel)

	// Load configuration
	configLoader := &config.YAMLConfigLoader{}
	config, err := configLoader.Load(conf.configPath)
	if err != nil {
		logger.Error("failed to load configuration",
			"error", err)
		panic(err)
	}

	// Validate configuration
	if config == nil {
		logger.Error("configuration is nil")
		panic("configuration is nil")
	}
	if len(config.Sections) == 0 {
		logger.Error("no sections defined in configuration")
		panic("no sections defined in configuration")
	}

	// Create a new storage
	storage := NewStorage()

	// Create a new handler
	handler := NewHandler(storage, config, logger)

	// Create server
	srv := &http.Server{
		Addr:         ":" + conf.port,
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
		logger.Error("server forced to shutdown",
			"error", err)
		panic(err)
	}

	logger.Info("server exited properly")
}
