package config

import (
	"os"
	"testing"
)

func TestFromEnv(t *testing.T) {
	// Store original env vars to restore them after the test
	originalPort := os.Getenv("UNIMOCK_PORT")
	originalLogLevel := os.Getenv("UNIMOCK_LOG_LEVEL")
	originalConfigPath := os.Getenv("UNIMOCK_CONFIG")
	defer func() {
		os.Setenv("UNIMOCK_PORT", originalPort)
		os.Setenv("UNIMOCK_LOG_LEVEL", originalLogLevel)
		os.Setenv("UNIMOCK_CONFIG", originalConfigPath)
	}()

	tests := []struct {
		name               string
		envVars            map[string]string
		expectedPort       string
		expectedLogLevel   string
		expectedConfigPath string
	}{
		{
			name:               "Default values when no env vars",
			envVars:            map[string]string{},
			expectedPort:       "8080",
			expectedLogLevel:   "info",
			expectedConfigPath: "config.yaml",
		},
		{
			name: "Custom port",
			envVars: map[string]string{
				"UNIMOCK_PORT": "9000",
			},
			expectedPort:       "9000",
			expectedLogLevel:   "info",
			expectedConfigPath: "config.yaml",
		},
		{
			name: "Custom log level",
			envVars: map[string]string{
				"UNIMOCK_LOG_LEVEL": "debug",
			},
			expectedPort:       "8080",
			expectedLogLevel:   "debug",
			expectedConfigPath: "config.yaml",
		},
		{
			name: "Custom config path",
			envVars: map[string]string{
				"UNIMOCK_CONFIG": "custom-config.yaml",
			},
			expectedPort:       "8080",
			expectedLogLevel:   "info",
			expectedConfigPath: "custom-config.yaml",
		},
		{
			name: "Invalid log level falls back to default",
			envVars: map[string]string{
				"UNIMOCK_LOG_LEVEL": "invalid",
			},
			expectedPort:       "8080",
			expectedLogLevel:   "info",
			expectedConfigPath: "config.yaml",
		},
		{
			name: "Case insensitive log level",
			envVars: map[string]string{
				"UNIMOCK_LOG_LEVEL": "DEBUG",
			},
			expectedPort:       "8080",
			expectedLogLevel:   "debug",
			expectedConfigPath: "config.yaml",
		},
		{
			name: "All custom values",
			envVars: map[string]string{
				"UNIMOCK_PORT":      "8081",
				"UNIMOCK_LOG_LEVEL": "warn",
				"UNIMOCK_CONFIG":    "test-config.yaml",
			},
			expectedPort:       "8081",
			expectedLogLevel:   "warn",
			expectedConfigPath: "test-config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant environment variables
			os.Unsetenv("UNIMOCK_PORT")
			os.Unsetenv("UNIMOCK_LOG_LEVEL")
			os.Unsetenv("UNIMOCK_CONFIG")

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call the function
			cfg := FromEnv()

			// Check results
			if cfg.Port != tt.expectedPort {
				t.Errorf("Expected Port %s, got %s", tt.expectedPort, cfg.Port)
			}
			if cfg.LogLevel != tt.expectedLogLevel {
				t.Errorf("Expected LogLevel %s, got %s", tt.expectedLogLevel, cfg.LogLevel)
			}
			if cfg.ConfigPath != tt.expectedConfigPath {
				t.Errorf("Expected ConfigPath %s, got %s", tt.expectedConfigPath, cfg.ConfigPath)
			}
		})
	}
}
