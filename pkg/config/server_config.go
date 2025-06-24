package config

import (
	"os"
	"strings"
)

// ServerConfig holds the basic server configuration options
// for controlling how the Unimock HTTP server operates.
type ServerConfig struct {
	// Port to listen on (default: "8080")
	// This controls which TCP port the HTTP server will bind to
	Port string `yaml:"port" json:"port"`

	// Log level: "debug", "info", "warn", "error" (default: "info")
	// Controls the verbosity of logs:
	// - debug: Most verbose, includes detailed debugging information
	// - info: Standard operational information
	// - warn: Only warning and error messages
	// - error: Only error messages
	LogLevel string `yaml:"log_level" json:"log_level"`

	// Path to configuration file (default: "config.yaml")
	// This controls where the mock configuration YAML file is located
	ConfigPath string `yaml:"config_path" json:"config_path"`

	// Path to scenarios file (optional)
	// If specified, scenarios will be loaded from this YAML file at server startup
	// If empty, only runtime API scenarios will be available
	ScenariosFile string `yaml:"scenarios_file" json:"scenarios_file"`
}

// NewDefaultServerConfig creates a ServerConfig with default values
// of port 8080 and info log level.
func NewDefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:       "8080",
		LogLevel:   "info",
		ConfigPath: "config.yaml",
	}
}

// FromEnv creates a ServerConfig from environment variables.
// It reads:
// - UNIMOCK_PORT: Port to listen on (default: "8080")
// - UNIMOCK_LOG_LEVEL: Log level (default: "info")
// - UNIMOCK_CONFIG: Path to configuration file (default: "config.yaml")
// - UNIMOCK_SCENARIOS_FILE: Path to scenarios file (optional)
//
// If an environment variable is not set, the default value is used.
func FromEnv() *ServerConfig {
	cfg := NewDefaultServerConfig()

	if port := os.Getenv("UNIMOCK_PORT"); port != "" {
		cfg.Port = port
	}

	if logLevel := os.Getenv("UNIMOCK_LOG_LEVEL"); logLevel != "" {
		// Normalize log level to lowercase
		logLevel = strings.ToLower(logLevel)
		// Only accept valid log levels
		if logLevel == "debug" || logLevel == "info" || logLevel == "warn" || logLevel == "error" {
			cfg.LogLevel = logLevel
		}
	}

	if configPath := os.Getenv("UNIMOCK_CONFIG"); configPath != "" {
		cfg.ConfigPath = configPath
	}

	if scenariosFile := os.Getenv("UNIMOCK_SCENARIOS_FILE"); scenariosFile != "" {
		cfg.ScenariosFile = scenariosFile
	}

	return cfg
}
