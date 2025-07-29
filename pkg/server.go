package pkg

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/router"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
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

// validateConfiguration validates that we have either sections or scenarios configured
func validateConfiguration(
	uniConfig *config.UniConfig,
	logger *slog.Logger,
) *ConfigError {
	if uniConfig == nil {
		err := "uni configuration is nil"
		logger.Error(err)
		return &ConfigError{Message: err}
	}

	sectionCount := len(uniConfig.Sections)
	scenarioCount := len(uniConfig.Scenarios)

	if sectionCount == 0 && scenarioCount == 0 {
		err := "no sections or scenarios defined in configuration"
		logger.Error(err)
		return &ConfigError{Message: err}
	}

	if sectionCount == 0 {
		logger.Info("running in scenarios-only mode - no sections configured")
	}

	return nil
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
// uniConfig must be non-nil and contain at least one section or scenario.
//
// Usage examples:
//
// 1. Using environment variables and file config:
//
//	// Load server configuration from environment variables
//	// UNIMOCK_PORT and UNIMOCK_LOG_LEVEL
//	serverConfig := config.FromEnv()
//
//	// Load unified configuration from file
//	uniConfig, err := config.LoadFromYAML(serverConfig.ConfigPath)
//	if err != nil {
//	    log.Fatal(err)
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
//	// Create unified configuration
//	uniConfig := &config.UniConfig{
//	    Sections: map[string]config.Section{
//	        "users": {
//	            PathPattern:  "/users/*",
//	            BodyIDPaths:  []string{"/id"},
//	            HeaderIDNames: []string{"X-User-ID"},
//	        },
//	    },
//	    Scenarios: []config.ScenarioConfig{
//	        {
//	            Method: "GET",
//	            Path: "/api/test",
//	            StatusCode: 200,
//	            Data: `{"message": "test"}`,
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

	// Validate uni configuration
	if err := validateConfiguration(uniConfig, logger); err != nil {
		return nil, err
	}

	// Create a new storage
	store := storage.NewUniStorage()

	// Create a new scenario storage
	scenarioStore := storage.NewScenarioStorage()

	// Create services
	uniService := service.NewUniService(store, uniConfig)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	// Load scenarios from uni config directly
	loadScenariosFromUniConfig(uniConfig, scenarioService, logger)

	// Create handlers with services
	uniHandler := handler.NewUniHandler(uniService, scenarioService, logger, uniConfig)
	scenarioHandler := handler.NewScenarioHandler(scenarioService, logger)
	techHandler := handler.NewTechHandler(techService, logger)

	// Create a router
	appRouter := router.NewRouter(
		uniHandler, techHandler, scenarioHandler,
		scenarioService, techService, logger, uniConfig,
	)

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

// loadScenariosFromUniConfig loads scenarios from the provided uni configuration
func loadScenariosFromUniConfig(
	uniConfig *config.UniConfig,
	scenarioService *service.ScenarioService,
	logger *slog.Logger,
) {
	// Check if uni config has scenarios
	if len(uniConfig.Scenarios) == 0 {
		logger.Debug("no scenarios found in uni config")
		return
	}

	logger.Info("loading scenarios from uni config", "count", len(uniConfig.Scenarios))

	// Convert to model scenarios and load them
	modelScenarios := make([]model.Scenario, 0, len(uniConfig.Scenarios))
	for _, sf := range uniConfig.Scenarios {
		modelScenarios = append(modelScenarios, sf.ToModelScenario())
	}
	ctx := context.Background()
	loadedCount := 0

	for _, scenario := range modelScenarios {
		_, err := scenarioService.CreateScenario(ctx, scenario)
		if err != nil {
			logger.Warn("failed to load scenario from uni config",
				"scenario_path", scenario.RequestPath,
				"scenario_uuid", scenario.UUID,
				"error", err)
			// Continue loading other scenarios even if one fails
			continue
		}
		loadedCount++
	}

	logger.Info("scenarios loaded from uni config",
		"total_scenarios", len(modelScenarios),
		"loaded_scenarios", loadedCount,
		"failed_scenarios", len(modelScenarios)-loadedCount)
}
