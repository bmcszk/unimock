package pkg

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/router"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
)

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

// setupLogger creates a new logger with the specified level
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

// NewServer initializes a new HTTP server with the provided configurations.
// Both serverConfig and mockConfig parameters are required.
//
// If serverConfig is nil, default values will be used:
//   - Port: "8080"
//   - LogLevel: "info"
//
// MockConfig must be non-nil and contain at least one section.
//
// Usage examples:
//
// 1. Using environment variables:
//
//	// Load server configuration from environment variables
//	// UNIMOCK_PORT and UNIMOCK_LOG_LEVEL
//	serverConfig := config.FromEnv()
//
//	// Create mock configuration
//	mockConfig := config.NewMockConfig()
//	mockConfig.Sections["users"] = config.Section{
//	    PathPattern:  "/users/*",
//	    BodyIDPaths:  []string{"/id"},
//	    HeaderIDName: "X-User-ID",
//	}
//
//	// Initialize server
//	srv, err := pkg.NewServer(serverConfig, mockConfig)
//
// 2. Using direct configuration:
//
//	// Create server configuration
//	serverConfig := &config.ServerConfig{
//	    Port:     "8080",
//	    LogLevel: "debug",
//	}
//
//	// Create mock configuration
//	mockConfig := &config.MockConfig{
//	    Sections: map[string]config.Section{
//	        "users": {
//	            PathPattern:  "/users/*",
//	            BodyIDPaths:  []string{"/id"},
//	            HeaderIDName: "X-User-ID",
//	        },
//	    },
//	}
//
//	// Initialize server
//	srv, err := pkg.NewServer(serverConfig, mockConfig)
//
// For a complete server setup:
//
//	srv, err := pkg.NewServer(serverConfig, mockConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start the server
//	log.Printf("Listening on %s", srv.Addr)
//	if err := srv.ListenAndServe(); err != nil {
//	    log.Fatal(err)
//	}
func NewServer(serverConfig *config.ServerConfig, mockConfig *config.MockConfig) (*http.Server, error) {
	if serverConfig == nil {
		serverConfig = config.NewDefaultServerConfig()
	}

	// Setup logger based on the log level from server config
	logger := setupLogger(serverConfig.LogLevel)

	logger.Info("initializing server",
		"port", serverConfig.Port,
		"log_level", serverConfig.LogLevel)

	// Validate mock configuration
	if mockConfig == nil {
		err := "mock configuration is nil"
		logger.Error(err)
		return nil, &ConfigError{Message: err}
	}
	if len(mockConfig.Sections) == 0 {
		err := "no sections defined in mock configuration"
		logger.Error(err)
		return nil, &ConfigError{Message: err}
	}

	// Create a new storage
	store := storage.NewMockStorage()

	// Create a new scenario storage
	scenarioStore := storage.NewScenarioStorage()

	// Create services
	mockService := service.NewMockService(store, mockConfig)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	// Create handlers with services
	mockHandler := handler.NewMockHandler(mockService, scenarioService, logger, mockConfig)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)
	techHandler := handler.NewTechHandler(techService, logger)

	// Create a router
	appRouter := router.NewRouter(mockHandler, techHandler, scenarioHandler, scenarioService, logger, mockConfig)

	// Create server
	srv := &http.Server{
		Addr:         ":" + serverConfig.Port,
		Handler:      appRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Return the created server
	logger.Info("server initialization complete, ready to start")
	return srv, nil
}
