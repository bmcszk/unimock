package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bmcszk/unimock/internal/config"
	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/storage"
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

// routeHandler is a http.Handler that routes requests to the appropriate handler based on path prefix
type routeHandler struct {
	mainHandler     http.Handler
	techHandler     http.Handler
	scenarioHandler http.Handler
	logger          *slog.Logger
}

// ServeHTTP implements the http.Handler interface
func (h *routeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/_uni/scenarios") {
		h.logger.Debug("routing to scenario handler", "path", r.URL.Path)
		h.scenarioHandler.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/_uni/") {
		h.logger.Debug("routing to technical handler", "path", r.URL.Path)
		h.techHandler.ServeHTTP(w, r)
		return
	}

	h.logger.Debug("routing to main handler", "path", r.URL.Path)
	h.mainHandler.ServeHTTP(w, r)
}

func main() {
	// Parse configuration
	envs := parseConfig()

	// Setup logger
	logger := setupLogger(envs.logLevel)
	logger.Info("starting server",
		"port", envs.port,
		"config_path", envs.configPath,
		"log_level", envs.logLevel)

	// Load configuration
	configLoader := &config.YAMLConfigLoader{}
	cfg, err := configLoader.Load(envs.configPath)
	if err != nil {
		logger.Error("failed to load configuration",
			"error", err)
		panic(err)
	}

	// Validate configuration
	if cfg == nil {
		logger.Error("configuration is nil")
		panic("configuration is nil")
	}
	if len(cfg.Sections) == 0 {
		logger.Error("no sections defined in configuration")
		panic("no sections defined in configuration")
	}

	// Create a new storage
	store := storage.NewStorage()

	// Create a new main handler
	mainHandler := handler.NewHandler(store, cfg, logger)

	// Create a new tech handler
	startTime := time.Now()
	techHandler := handler.NewTechHandler(logger, startTime)

	// Create a new scenario handler
	scenarioHandler := handler.NewScenarioHandler(store, logger)

	// Create a router handler
	router := &routeHandler{
		mainHandler:     mainHandler,
		techHandler:     techHandler,
		scenarioHandler: scenarioHandler,
		logger:          logger,
	}

	// Create server
	srv := &http.Server{
		Addr:         ":" + envs.port,
		Handler:      router,
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
