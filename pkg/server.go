package pkg

import (
	"context"
	"fmt"
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

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
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
	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(jsonHandler)
}

// NewServer initializes a new HTTP server with the provided configurations.
// Both serverConfig and uniConfig parameters are required.
//
// If serverConfig is nil, default values will be used:
//   - Port: "8080"
//   - LogLevel: "info"
//
// UniConfig must be non-nil and contain at least one section.
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
//	uniConfig := config.NewUniConfig()
//	uniConfig.Sections["users"] = config.Section{
//	    PathPattern:  "/users/*",
//	    BodyIDPaths:  []string{"/id"},
//	    HeaderIDName: "X-User-ID",
//	}
//
//	// Initialize server
//	srv, err := pkg.NewServer(serverConfig, uniConfig)
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
//	uniConfig := &config.UniConfig{
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
//	srv, err := pkg.NewServer(serverConfig, uniConfig)
//
// For a complete server setup:
//
//	srv, err := pkg.NewServer(serverConfig, uniConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start the server
//	log.Printf("Listening on %s", srv.Addr)
//	if err := srv.ListenAndServe(); err != nil {
//	    log.Fatal(err)
//	}
func NewServer(serverConfig *config.ServerConfig, uniConfig *config.UniConfig) (*http.Server, error) {
	if serverConfig == nil {
		serverConfig = config.NewDefaultServerConfig()
	}

	// Setup logger based on the log level from server config
	logger := setupLogger(serverConfig.LogLevel)

	logger.Info("initializing server",
		"port", serverConfig.Port,
		"log_level", serverConfig.LogLevel)

	// Validate mock configuration
	if uniConfig == nil {
		err := "mock configuration is nil"
		logger.Error(err)
		return nil, &ConfigError{Message: err}
	}
	if len(uniConfig.Sections) == 0 {
		err := "no sections defined in mock configuration"
		logger.Error(err)
		return nil, &ConfigError{Message: err}
	}

	// Create a new storage
	store := storage.NewUniStorage()

	// Create a new scenario storage
	scenarioStore := storage.NewScenarioStorage()

	// Create services
	uniService := service.NewUniService(store, uniConfig)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	// Load scenarios from file if specified
	if err := loadScenariosFromFile(serverConfig, scenarioService, logger); err != nil {
		logger.Error("failed to load scenarios from file", "error", err)
		return nil, &ConfigError{Message: err.Error()}
	}

	// Create handlers with services
	mockHandler := handler.NewUniHandler(uniService, scenarioService, techService, logger, uniConfig)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)
	techHandler := handler.NewTechHandler(techService, logger)

	// Create a router
	appRouter := router.NewRouter(mockHandler, techHandler, scenarioHandler, scenarioService, logger, uniConfig)

	// Create server
	srv := &http.Server{
		Addr:         ":" + serverConfig.Port,
		Handler:      appRouter,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Return the created server
	logger.Info("server initialization complete, ready to start")
	return srv, nil
}

// loadScenariosFromFile loads scenarios from the configured scenarios file
func loadScenariosFromFile(
	serverConfig *config.ServerConfig, 
	scenarioService *service.ScenarioService, 
	logger *slog.Logger,
) error {
	if serverConfig.ScenariosFile == "" {
		logger.Debug("no scenarios file configured, skipping file-based scenarios loading")
		return nil
	}

	logger.Info("loading scenarios from file", "file", serverConfig.ScenariosFile)

	// Load scenarios configuration from file
	scenariosConfig, err := config.LoadScenariosFromYAML(serverConfig.ScenariosFile)
	if err != nil {
		return fmt.Errorf("failed to load scenarios from file %s: %w", serverConfig.ScenariosFile, err)
	}

	// Convert to model scenarios and load them
	modelScenarios := scenariosConfig.ToModelScenarios()
	ctx := context.Background()
	loadedCount := 0

	for _, scenario := range modelScenarios {
		if err := scenarioService.CreateScenario(ctx, scenario); err != nil {
			logger.Warn("failed to load scenario from file", 
				"scenario_path", scenario.RequestPath,
				"scenario_uuid", scenario.UUID,
				"error", err)
			// Continue loading other scenarios even if one fails
			continue
		}
		loadedCount++
	}

	logger.Info("scenarios loaded from file", 
		"file", serverConfig.ScenariosFile,
		"total_scenarios", len(modelScenarios),
		"loaded_scenarios", loadedCount,
		"failed_scenarios", len(modelScenarios)-loadedCount)

	return nil
}
