/*
Package config provides configuration structures for the Unimock server.

The configuration is split into two main parts:

1. ServerConfig: Controls server behavior like port and logging level
2. MockConfig: Defines mock behavior for API endpoints

# Server Configuration

ServerConfig can be created in several ways:

1. From environment variables using FromEnv():

	// Load server configuration from environment variables:
	// - UNIMOCK_PORT (default: "8080")
	// - UNIMOCK_LOG_LEVEL (default: "info")
	// - UNIMOCK_CONFIG (default: "config.yaml")
	serverConfig := config.FromEnv()

2. Using default values:

	// Create with defaults:
	// - Port: "8080"
	// - LogLevel: "info"
	// - ConfigPath: "config.yaml"
	serverConfig := config.NewDefaultServerConfig()

3. Creating directly:

	// Create with custom values
	serverConfig := &config.ServerConfig{
		Port:       "8081",
		LogLevel:   "debug",
		ConfigPath: "custom-config.yaml",
	}

# Mock Configuration

MockConfig can be created in two ways:

1. Loading from a YAML file:

	// Load from a YAML file
	mockConfig, err := config.LoadFromYAML("config.yaml")
	if err != nil {
		// Handle error
	}

2. Creating directly in code:

	// Create empty configuration
	mockConfig := config.NewMockConfig()

	// Add sections
	mockConfig.Sections["users"] = config.Section{
		PathPattern:  "/users/*",
		BodyIDPaths:  []string{"/id", "/data/id"},
		HeaderIDName: "X-User-ID",
	}

# Complete Usage Example

	// Load server configuration from environment variables
	serverConfig := config.FromEnv()

	// Create mock configuration in code
	mockConfig := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:  "/users/*",
				BodyIDPaths:  []string{"/id", "/data/id"},
				HeaderIDName: "X-User-ID",
			},
		},
	}

	// Initialize server with configurations
	srv, err := pkg.NewServer(serverConfig, mockConfig)

# Configuration Details

The MockConfig structure supports advanced patterns for extracting IDs from:
- URL paths (using wildcard patterns)
- Request headers (using header names)
- JSON/XML bodies (using XPath-like expressions)

This makes it flexible enough to work with virtually any HTTP-based API.

See the example directory for a complete working example:
https://github.com/bmcszk/unimock/tree/master/example
*/
package config
