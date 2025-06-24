package config_test

import (
	"os"
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
)

const (
	defaultPort       = "8080"
	defaultLogLevel   = "info"
	defaultConfigPath = "config.yaml"
	envLogLevel       = "UNIMOCK_LOG_LEVEL"
)

func TestFromEnv(t *testing.T) {
	t.Run("DefaultValues", testDefaultValues)
	t.Run("CustomValues", testCustomValues)
	t.Run("InvalidValues", testInvalidValues)
}

func testDefaultValues(t *testing.T) {
	restoreEnv := setupEnvTest(t)
	defer restoreEnv()

	clearEnvVars()
	cfg := config.FromEnv()

	if cfg.Port != defaultPort {
		t.Errorf("Expected Port %s, got %s", defaultPort, cfg.Port)
	}
	if cfg.LogLevel != defaultLogLevel {
		t.Errorf("Expected LogLevel %s, got %s", defaultLogLevel, cfg.LogLevel)
	}
	if cfg.ConfigPath != defaultConfigPath {
		t.Errorf("Expected ConfigPath %s, got %s", defaultConfigPath, cfg.ConfigPath)
	}
	if cfg.ScenariosFile != "" {
		t.Errorf("Expected ScenariosFile to be empty, got %s", cfg.ScenariosFile)
	}
}

func testCustomValues(t *testing.T) {
	restoreEnv := setupEnvTest(t)
	defer restoreEnv()

	tests := []struct {
		name string
		env  map[string]string
		expected config.ServerConfig
	}{
		{
			name: "Custom port",
			env: map[string]string{"UNIMOCK_PORT": "9000"},
			expected: config.ServerConfig{Port: "9000", LogLevel: "info", ConfigPath: "config.yaml", ScenariosFile: ""},
		},
		{
			name: "Custom log level",
			env: map[string]string{"UNIMOCK_LOG_LEVEL": "debug"},
			expected: config.ServerConfig{
				Port: "8080", LogLevel: "debug", ConfigPath: "config.yaml", ScenariosFile: "",
			},
		},
		{
			name: "Custom scenarios file",
			env: map[string]string{"UNIMOCK_SCENARIOS_FILE": "test-scenarios.yaml"},
			expected: config.ServerConfig{
				Port: "8080", LogLevel: "info", ConfigPath: "config.yaml", 
				ScenariosFile: "test-scenarios.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			setEnvVars(tt.env)
			cfg := config.FromEnv()
			validateConfig(t, cfg, tt.expected)
		})
	}
}

func testInvalidValues(t *testing.T) {
	restoreEnv := setupEnvTest(t)
	defer restoreEnv()

	clearEnvVars()
	_ = os.Setenv("UNIMOCK_LOG_LEVEL", "invalid")
	cfg := config.FromEnv()

	if cfg.LogLevel != defaultLogLevel {
		t.Errorf("Expected LogLevel to fallback to %s, got %s", defaultLogLevel, cfg.LogLevel)
	}
}

func setupEnvTest(_ *testing.T) func() {
	originalPort := os.Getenv("UNIMOCK_PORT")
	originalLogLevel := os.Getenv(envLogLevel)
	originalConfigPath := os.Getenv("UNIMOCK_CONFIG")
	originalScenariosFile := os.Getenv("UNIMOCK_SCENARIOS_FILE")
	
	return func() {
		_ = os.Setenv("UNIMOCK_PORT", originalPort)
		_ = os.Setenv(envLogLevel, originalLogLevel)
		_ = os.Setenv("UNIMOCK_CONFIG", originalConfigPath)
		_ = os.Setenv("UNIMOCK_SCENARIOS_FILE", originalScenariosFile)
	}
}

func clearEnvVars() {
	os.Unsetenv("UNIMOCK_PORT")
	os.Unsetenv(envLogLevel)
	os.Unsetenv("UNIMOCK_CONFIG")
	os.Unsetenv("UNIMOCK_SCENARIOS_FILE")
}

func setEnvVars(envVars map[string]string) {
	for k, v := range envVars {
		_ = os.Setenv(k, v)
	}
}

func validateConfig(t *testing.T, actual *config.ServerConfig, expected config.ServerConfig) {
	t.Helper()
	if actual.Port != expected.Port {
		t.Errorf("Expected Port %s, got %s", expected.Port, actual.Port)
	}
	if actual.LogLevel != expected.LogLevel {
		t.Errorf("Expected LogLevel %s, got %s", expected.LogLevel, actual.LogLevel)
	}
	if actual.ConfigPath != expected.ConfigPath {
		t.Errorf("Expected ConfigPath %s, got %s", expected.ConfigPath, actual.ConfigPath)
	}
	if actual.ScenariosFile != expected.ScenariosFile {
		t.Errorf("Expected ScenariosFile %s, got %s", expected.ScenariosFile, actual.ScenariosFile)
	}
}
