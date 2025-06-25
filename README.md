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
- **File-Based Scenarios**: Load predefined scenarios from YAML files at startup

## Installation

```bash
go get github.com/bmcszk/unimock
```

## Usage Options

Unimock can be used in three ways:

### 1. As a Standalone Application

Run the mock server directly:

```bash
# Using environment variables for configuration
UNIMOCK_PORT=8080 UNIMOCK_CONFIG=config.yaml go run main.go

# With predefined scenarios file
UNIMOCK_PORT=8080 UNIMOCK_CONFIG=config.yaml UNIMOCK_SCENARIOS_FILE=scenarios.yaml go run main.go
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
// Optional: Set scenarios file path
serverConfig.ScenariosFile = "scenarios.yaml"

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

### 3. As a Docker Container

Unimock is available as a Docker image from GitHub Container Registry:

#### Quick Start with Docker

```bash
# Run with default configuration
docker run -p 8080:8080 ghcr.io/bmcszk/unimock:latest

# Run with custom configuration
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/unimock/config.yaml \
  ghcr.io/bmcszk/unimock:latest

# Run with scenarios file
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/unimock/config.yaml \
  -v $(pwd)/scenarios.yaml:/etc/unimock/scenarios.yaml \
  -e UNIMOCK_SCENARIOS_FILE=/etc/unimock/scenarios.yaml \
  ghcr.io/bmcszk/unimock:latest
```

#### Available Tags

- `latest` - Latest stable release (from latest version tag)
- `v1.x.x` - Specific version tags (e.g., `v1.2.0`)
- `v1.x` - Minor version tags (e.g., `v1.2`)
- `v1` - Major version tags (e.g., `v1`)

#### Docker Compose Example

```yaml
version: '3.8'
services:
  unimock:
    image: ghcr.io/bmcszk/unimock:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/etc/unimock/config.yaml
      - ./scenarios.yaml:/etc/unimock/scenarios.yaml
    environment:
      - UNIMOCK_PORT=8080
      - UNIMOCK_CONFIG=/etc/unimock/config.yaml
      - UNIMOCK_SCENARIOS_FILE=/etc/unimock/scenarios.yaml
      - UNIMOCK_LOG_LEVEL=info
    healthcheck:
      test: ["CMD-SHELL", "wget --quiet --tries=1 --spider http://localhost:8080/_uni/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

#### Building Custom Image

```bash
# Clone and build locally
git clone https://github.com/bmcszk/unimock.git
cd unimock
docker build -t unimock:local .

# Run your custom build
docker run -p 8080:8080 unimock:local
```

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
- `UNIMOCK_SCENARIOS_FILE` - The path to scenarios YAML file for predefined scenarios (optional)

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

## Scenarios File Configuration

Unimock supports loading predefined scenarios from a YAML file at server startup. This feature allows you to define static mock responses that will be available immediately when the server starts, working alongside the dynamic runtime API scenarios.

### Configuration

Set the scenarios file path using the environment variable:

```bash
export UNIMOCK_SCENARIOS_FILE=scenarios.yaml
```

Or when using Unimock as a library:

```go
serverConfig := config.FromEnv()
serverConfig.ScenariosFile = "scenarios.yaml"
```

### YAML File Format

Create a scenarios file with the following structure:

```yaml
scenarios:
  - uuid: "get-user-123"                    # Optional: auto-generated if omitted
    method: "GET"                           # Required: HTTP method
    path: "/api/users/123"                  # Required: URL path
    status_code: 200                        # Optional: default 200
    content_type: "application/json"        # Optional: default "application/json"
    data: |                                 # Optional: response body
      {
        "id": "123",
        "name": "John Doe",
        "email": "john@example.com"
      }
    headers:                                # Optional: additional headers
      X-Custom-Header: "custom-value"
      
  - method: "POST"
    path: "/api/users"
    status_code: 201
    location: "/api/users/456"              # Optional: Location header for redirects
    content_type: "application/json"
    data: |
      {
        "id": "456",
        "message": "User created successfully"
      }
      
  - method: "HEAD"
    path: "/api/health"
    status_code: 200
    # HEAD requests typically don't include response bodies
```

### Key Features

- **Optional Configuration**: If no scenarios file is specified, the server runs normally with only runtime API scenarios
- **Validation**: The YAML file is validated at startup with clear error messages
- **Runtime Integration**: File-based scenarios work alongside runtime API scenarios
- **Flexible Structure**: Support for all HTTP methods, custom headers, and various content types
- **Auto-Generated UUIDs**: Scenarios without UUIDs get auto-generated identifiers

### Example Usage

1. Create a scenarios file:

```bash
cat > scenarios.yaml << EOF
scenarios:
  - method: "GET"
    path: "/api/status"
    status_code: 200
    data: '{"status": "healthy", "timestamp": "2024-01-01T00:00:00Z"}'
    
  - method: "GET"  
    path: "/api/users/test-user"
    status_code: 200
    data: '{"id": "test-user", "name": "Test User", "role": "tester"}'
EOF
```

2. Start the server:

```bash
UNIMOCK_SCENARIOS_FILE=scenarios.yaml make run
```

3. Test the predefined scenarios:

```bash
curl http://localhost:8080/api/status
curl http://localhost:8080/api/users/test-user
```

The scenarios will be available immediately when the server starts, and you can still add more scenarios dynamically using the runtime API.

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
