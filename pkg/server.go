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
	unifiedConfig *config.UnifiedConfig,
	logger *slog.Logger,
) *ConfigError {
	if unifiedConfig == nil {
		err := "unified configuration is nil"
		logger.Error(err)
		return &ConfigError{Message: err}
	}

	sectionCount := len(unifiedConfig.Sections)
	scenarioCount := len(unifiedConfig.Scenarios)

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
// Both serverConfig and unifiedConfig parameters are required.
//
// If serverConfig is nil, default values will be used:
//   - Port: "8080"
//   - LogLevel: "info"
//
// UnifiedConfig must be non-nil and contain at least one section or scenario.
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
//	unifiedConfig, err := config.LoadUnifiedFromYAML(serverConfig.ConfigPath)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Initialize server
//	srv, err := pkg.NewServer(serverConfig, unifiedConfig)
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
//	unifiedConfig := &config.UnifiedConfig{
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
//	srv, err := pkg.NewServer(serverConfig, unifiedConfig)
//
// For a complete server setup:
//
//	srv, err := pkg.NewServer(serverConfig, unifiedConfig)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start the server
//	log.Printf("Listening on %s", srv.Addr)
//	if err := srv.ListenAndServe(); err != nil {
//	    log.Fatal(err)
//	}
func NewServer(serverConfig *config.ServerConfig, unifiedConfig *config.UnifiedConfig) (*http.Server, error) {
	if serverConfig == nil {
		serverConfig = config.NewDefaultServerConfig()
	}

	// Setup logger based on the log level from server config
	logger := setupLogger(serverConfig.LogLevel)

	logger.Info("initializing server",
		"port", serverConfig.Port,
		"log_level", serverConfig.LogLevel)

	// Validate unified configuration
	if err := validateConfiguration(unifiedConfig, logger); err != nil {
		return nil, err
	}

	// Create a new storage
	store := storage.NewUniStorage()

	// Create a new scenario storage
	scenarioStore := storage.NewScenarioStorage()

	// Convert unified config to uni config for backward compatibility
	uniConfig := &config.UniConfig{
		Sections: unifiedConfig.Sections,
	}

	// Create services
	uniService := service.NewUniService(store, uniConfig)
	scenarioService := service.NewScenarioService(scenarioStore)
	techService := service.NewTechService(time.Now())

	// Load scenarios from unified config directly
	loadScenariosFromUnifiedConfig(unifiedConfig, scenarioService, logger)

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

// loadScenariosFromUnifiedConfig loads scenarios from the provided unified configuration
func loadScenariosFromUnifiedConfig(
	unifiedConfig *config.UnifiedConfig,
	scenarioService *service.ScenarioService,
	logger *slog.Logger,
) {
	// Check if unified config has scenarios
	if len(unifiedConfig.Scenarios) == 0 {
		logger.Debug("no scenarios found in unified config")
		return
	}

	logger.Info("loading scenarios from unified config", "count", len(unifiedConfig.Scenarios))

	// Convert to model scenarios and load them
	modelScenarios := make([]model.Scenario, 0, len(unifiedConfig.Scenarios))
	for _, sf := range unifiedConfig.Scenarios {
		modelScenarios = append(modelScenarios, sf.ToModelScenario())
	}
	ctx := context.Background()
	loadedCount := 0

	for _, scenario := range modelScenarios {
		_, err := scenarioService.CreateScenario(ctx, scenario)
		if err != nil {
			logger.Warn("failed to load scenario from unified config",
				"scenario_path", scenario.RequestPath,
				"scenario_uuid", scenario.UUID,
				"error", err)
			// Continue loading other scenarios even if one fails
			continue
		}
		loadedCount++
	}

	logger.Info("scenarios loaded from unified config",
		"total_scenarios", len(modelScenarios),
		"loaded_scenarios", loadedCount,
		"failed_scenarios", len(modelScenarios)-loadedCount)
}
