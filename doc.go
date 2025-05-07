/*
Package main provides the entry point for running Unimock as a standalone application.

Unimock is a universal mock HTTP server for end-to-end testing that can be used in two ways:

1. As a standalone application (using this package)
2. As a library in your own Go code (using the pkg package)

# Standalone Usage

Run Unimock as a standalone application:

	go run github.com/bmcszk/unimock

	# Or after building
	./unimock

Configuration is loaded from environment variables:

	UNIMOCK_PORT=8080 UNIMOCK_LOG_LEVEL=debug UNIMOCK_CONFIG=config.yaml ./unimock

Environment variables:
  - UNIMOCK_PORT: HTTP port (default: "8080")
  - UNIMOCK_LOG_LEVEL: Log level: debug, info, warn, error (default: "info")
  - UNIMOCK_CONFIG: Path to YAML configuration file (default: "config.yaml")

# Library Usage

Import Unimock in your Go code:

	import (
		"github.com/bmcszk/unimock/pkg"
		"github.com/bmcszk/unimock/pkg/config"
	)

See the example directory for a complete working example:
https://github.com/bmcszk/unimock/tree/master/example

Basic usage:

	// Load configuration from environment variables
	serverConfig := config.FromEnv()

	// Load mock configuration from YAML file
	mockConfig, err := config.LoadFromYAML(serverConfig.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Or create mock configuration in code
	mockConfig := &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:  "/users/*",
				BodyIDPaths:  []string{"/id"},
				HeaderIDName: "X-User-ID",
			},
		},
	}

	// Initialize and start the server
	srv, err := pkg.NewServer(serverConfig, mockConfig)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	log.Printf("Server listening on port %s", serverConfig.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

# Configuration

See the pkg/config package documentation for details on configuration options:
https://pkg.go.dev/github.com/bmcszk/unimock/pkg/config
*/
package main
