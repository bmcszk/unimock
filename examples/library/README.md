# Unimock Example

This example demonstrates how to use Unimock as a library with in-code configuration.

## How It Works

The example shows:

1. Creating a mock configuration with API endpoint patterns
2. Setting up server configuration
3. Initializing and running the Unimock server
4. Handling graceful shutdown

## Running the Example

Run the example with:

```bash
go run .
```

This will start a mock server on port 8081. You can then make requests to the server:

```bash
# Create a user
curl -X POST -H "Content-Type: application/json" -d '{"id":"123","name":"Test User"}' http://localhost:8081/users

# Get the user
curl http://localhost:8081/users/123

# Update the user
curl -X PUT -H "Content-Type: application/json" -d '{"id":"123","name":"Updated User"}' http://localhost:8081/users/123

# Delete the user
curl -X DELETE http://localhost:8081/users/123
```

## Configuration Details

The example configures two API endpoints:

1. `/api/*` - For general API requests
2. `/users/*` - For user-specific requests

Each endpoint is configured with:
- Path pattern matching
- JSON body ID extraction paths
- Header-based ID extraction

## Custom Configuration

To customize the configuration, modify the `uniConfig` and `serverConfig` objects in `main.go`:

```go
// Server configuration
serverConfig := &config.ServerConfig{
    Port:       "8081",    // Change the port
    LogLevel:   "debug",   // Change log level: debug, info, warn, error
    ConfigPath: "config.yaml", // Required but not used for in-code config
}

// Mock configuration
uniConfig := &config.UniConfig{
    Sections: map[string]config.Section{
        "custom": {
            PathPattern:   "/custom/*",               // Custom path pattern
            BodyIDPaths:   []string{"/customId"},     // Custom ID paths
            HeaderIDNames: []string{"X-Custom-ID"},   // Custom headers (multiple supported)
        },
    },
}
``` 
