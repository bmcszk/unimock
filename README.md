# Unimock

A Universal Mock HTTP server designed for e2e testing scenarios where you need to mock third-party REST services. This server can work as a RESTful service, XML service, or any other HTTP-based service.

## Purpose

Unimock was created to solve a common problem in e2e testing: the need to mock third-party REST services that you don't control. Instead of building custom mock services for each test scenario, Unimock provides a universal solution that can handle any HTTP-based service.

## Key Features

- **Universal Support**: Works with any HTTP-based service (REST, XML, etc.)
- **Flexible ID Extraction**: Extracts IDs from headers, JSON/XML bodies, or paths
- **Enhanced Wildcard Patterns**: Support for single (*) and recursive (**) wildcards in path patterns
- **Strict Path Matching**: Optional strict mode for precise path and resource validation
- **Thread-Safe Storage**: In-memory storage with mutex protection
- **Full HTTP Support**: All HTTP methods (GET, HEAD, POST, PUT, DELETE)
- **Technical Endpoints**: Health check, metrics, and scenario management
- **Scenario-Based Mocking**: Create predefined scenarios for paths that bypass regular mock behavior

## Installation

```bash
go get github.com/bmcszk/unimock
```

## Usage Options

Unimock can be used in two ways:

### 1. As a Standalone Application

Run the mock server directly:

```bash
# Using environment variables for configuration
UNIMOCK_PORT=8080 UNIMOCK_CONFIG=config.yaml go run main.go
```

### 2. As a Library in Your Go Code

Import Unimock in your Go application:

```go
import (
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
)

// Load or create configuration
serverConfig := config.FromEnv()
mockConfig := &config.MockConfig{
    Sections: map[string]config.Section{
        "users": {
            PathPattern:  "/users/*",
            StrictPath:   false, // Optional: enable strict path matching
            ReturnBody:   true,  // Optional: return resource bodies in responses
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

// Start the server
if err := srv.ListenAndServe(); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

See the [example directory](./example) for a complete working example of using Unimock as a library.

## Quick Start

To run the mock server:

```bash
make run
```

Or directly:

```bash
go run main.go
```

This will start the server on port 8080.

## Development Commands

The project includes Makefile targets for common tasks:

```bash
# Standard development
make build      # Build the application
make test       # Run all tests
make run        # Build and run the application
make clean      # Clean build artifacts

# Kubernetes development
make kind-start # Start a local Kubernetes cluster with KinD
make helm-lint  # Lint the Helm chart
make tilt-run   # Run Tilt for local development in Kubernetes
make k8s-setup  # Deploy to Kubernetes using Helm
make kind-stop  # Delete the Kubernetes cluster
```

## Configuration

The server can be configured with environment variables:

- `UNIMOCK_PORT` - The port to listen on (default: `8080`)
- `UNIMOCK_CONFIG` - The path to the configuration file (default: `config.yaml`)
- `UNIMOCK_LOG_LEVEL` - The log level (default: `info`)

## Basic Usage

### Create a resource

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "test"}' \
  http://localhost:8080/users
```

### Retrieve a resource

```bash
curl -X GET http://localhost:8080/users/123
```

### Update a resource

```bash
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "updated"}' \
  http://localhost:8080/users/123
```

### Delete a resource

```bash
curl -X DELETE http://localhost:8080/users/123
```

### Check if a resource exists (HEAD)

```bash
curl -I http://localhost:8080/users/123
```

HEAD requests return the same headers and status code as GET requests but without the response body, making them perfect for checking resource existence or metadata.

## Advanced Path Matching

Unimock supports advanced path pattern matching with wildcards and strict validation modes.

### Wildcard Patterns

#### Single Wildcard (*)
Matches exactly one path segment:

```yaml
sections:
  users:
    path_pattern: "/users/*"        # Matches: /users/123, /users/abc
    # Does NOT match: /users/123/posts
```

#### Recursive Wildcard (**)
Matches zero or more path segments:

```yaml
sections:
  api_resources:
    path_pattern: "/api/**"         # Matches: /api, /api/v1, /api/v1/users/123/posts
```

#### Mixed Patterns
Combine both wildcard types:

```yaml
sections:
  user_posts:
    path_pattern: "/users/*/posts/**"  # Matches: /users/123/posts/456/comments
```

### Strict Path Matching

Control validation behavior with the `strict_path` flag:

```yaml
sections:
  flexible_api:
    path_pattern: "/users/**"
    strict_path: false              # Default: flexible matching
    
  strict_api:
    path_pattern: "/admin/**"
    strict_path: true               # Strict: exact path validation
```

### Response Body Control

Control whether POST/PUT/DELETE operations return resource bodies with the `return_body` flag:

```yaml
sections:
  minimal_api:
    path_pattern: "/api/minimal/**"
    return_body: false              # Default: empty response bodies
    
  full_api:
    path_pattern: "/api/full/**"
    return_body: true               # Return resource bodies in responses
```

#### Behavior Scenarios

# Scenario 1: Flexible Path Matching (strict_path=false)
```yaml
given:
- pattern "/users/**"
- POST /users/subpath body: { "id": 1 }
- strict_path: false

when:
- GET/PUT/DELETE /users/1

then:
✅ Operations succeed (cross-path access allowed)
```

# Scenario 2: Strict Path Matching (strict_path=true)
```yaml
given:
- pattern "/users/**"  
- POST /users/subpath body: { "id": 1 }
- strict_path: true

when:
- GET/PUT/DELETE /users/1
then:
❌ 404 Not Found (different path structure)

when:
- GET/PUT/DELETE /users/subpath/1
then:
✅ Operations succeed (same path structure)
```

**Key Points:**
- `strict_path=true` enforces path structure compatibility for resource access
- `strict_path=false` (default) allows flexible cross-path resource access via extracted IDs
- With `strict_path=true`, resources are only accessible via paths that extend their creation path
- Example: Resource created at `/users/subpath` accessible via `/users/subpath/123` but not `/users/123`
- PUT operations still support upsert when `strict_path=false`, return 404 when `strict_path=true` and resource doesn't exist

## Request/Response Transformations (Library Mode Only)

When using Unimock as a library, you can configure request and response transformations to modify data before storage or after retrieval. Transformations are applied using simple Go functions and are completely optional.

### Transformation Functions

Both request and response transformations use the same simple function signature:

```go
func(data *model.MockData) (*model.MockData, error)
```

- **Input**: MockData containing the request/response data
- **Output**: Modified MockData or an error
- **Error Handling**: Any error returns HTTP 500 Internal Server Error

### Configuration Example

```go
import (
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
    "github.com/bmcszk/unimock/pkg/model"
)

// Create transformation configuration
transformConfig := config.NewTransformationConfig()

// Add request transformation (applied before storing)
transformConfig.AddRequestTransform(func(data *model.MockData) (*model.MockData, error) {
    // Modify request data before storage
    modifiedData := *data
    modifiedData.Body = []byte(`{"id": "123", "name": "transformed", "request_processed": true}`)
    return &modifiedData, nil
})

// Add response transformation (applied after retrieval)
transformConfig.AddResponseTransform(func(data *model.MockData) (*model.MockData, error) {
    // Modify response data before sending
    modifiedData := *data
    modifiedData.Body = []byte(`{"id": "123", "name": "transformed", "response_processed": true}`)
    return &modifiedData, nil
})

// Apply to specific section
mockConfig := &config.MockConfig{
    Sections: map[string]config.Section{
        "users": {
            PathPattern:     "/users/*",
            StrictPath:      false, // Optional: enable strict path matching
            ReturnBody:      true,  // Optional: return resource bodies in responses
            BodyIDPaths:     []string{"/id"},
            CaseSensitive:   false,
            Transformations: transformConfig, // Apply transformations
        },
    },
}

server, err := pkg.NewServer(
    pkg.WithPort(8080),
    pkg.WithMockConfig(mockConfig),
)
```

### Key Features

- **Optional**: Both request and response transformations are completely optional
- **Library-Only**: Transformations cannot be configured via YAML files
- **Error Safety**: All transformation errors result in HTTP 500 responses
- **Panic Recovery**: Transformation panics are recovered and logged
- **Simple API**: Single function signature for both request and response transformations

### Use Cases

1. **Data Validation**: Validate incoming request data
2. **Data Enrichment**: Add computed fields or timestamps
3. **Format Conversion**: Convert between different data formats
4. **Security**: Add/remove sensitive information
5. **Testing**: Simulate different response scenarios

## Documentation

- [HTTP Methods](docs/http_methods.md) - How different HTTP methods are handled
- [ID Extraction](docs/id_extraction.md) - How IDs are extracted from requests
- [Configuration](docs/configuration.md) - How to configure the server
- [Technical Endpoints](docs/technical_endpoints.md) - Special endpoints for management
- [Usage Examples](docs/examples.md) - More detailed usage examples
- [Scenario System](docs/scenarios.md) - How to use and test with the scenario storage, including path-based mocking

## For Developers and Contributors

- [Storage System (Internal)](docs/storage.md) - Internal storage implementation details
- [Kubernetes Deployment](helm/unimock/README.md) - Deploy Unimock with Helm
- [Local Development with Tilt](tilt/README.md) - Run and develop with Tilt in Kubernetes

## Use Cases

1. **E2E Testing**: Mock third-party services in your test environment
2. **Development**: Simulate external services during development
3. **Integration Testing**: Test your application's integration with external services
4. **API Development**: Prototype and test API designs

## License

This project is licensed under the [MIT License](LICENSE).
