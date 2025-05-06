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

	"github.com/bmcszk/unimock/config"
)

type Envs struct {
	port       string
	configPath string
	logLevel   string
}

func parseConfig() *Envs {
	// TODO change flags to envs
	envs := &Envs{}
	flag.StringVar(&envs.port, "port", "8080", "Port to listen on")
	flag.StringVar(&envs.configPath, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&envs.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

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
