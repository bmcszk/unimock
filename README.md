# Unimock

A universal HTTP mock server for end-to-end testing. Mock any REST API, GraphQL service, or XML endpoint with flexible configuration and scenario-based responses.

## Quick Start

### 1. Run with Docker

```bash
docker run -p 8080:8080 ghcr.io/bmcszk/unimock:latest
```

### 2. Test the mock server

```bash
# Add test data
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "1", "name": "John Doe", "email": "john@example.com"}'

# Get the data back  
curl http://localhost:8080/api/users/1
```

### 3. Check health

```bash
curl http://localhost:8080/_uni/health
```

## What is Unimock?

Unimock is a testing tool that creates fake HTTP services. Instead of calling real APIs during testing, your application calls Unimock, which responds with test data you control.

**Use cases:**
- Test your app without depending on external services
- Create predictable test scenarios
- Simulate API errors and edge cases
- Speed up integration tests

## Key Features

- **Universal**: Works with REST, GraphQL, XML, or any HTTP service
- **Smart ID extraction**: Finds IDs in URLs, JSON, XML, or headers automatically
- **Scenarios**: Pre-defined responses for specific test cases
- **Thread-safe**: Handle multiple requests simultaneously
- **Easy configuration**: Simple YAML setup

## Configuration

Create a `config.yaml` file to define which endpoints to mock:

```yaml
sections:
  - path: "/api/users/*"     # Match /api/users/123
    id_path: "/id"           # Extract ID from JSON body (XPath-like syntax)
    return_body: true        # Return the posted data
  - path: "/api/orders/*"
    id_path: "/order_id"     # Extract order_id field
    header_id_name: "X-Order-ID"  # Get ID from header
```

**[📖 Full Configuration Guide](docs/configuration.md)**

## Scenarios

Pre-define specific responses for testing:

```yaml
scenarios:
  - uuid: "user-not-found"
    method: "GET"
    path: "/api/users/999"
    status_code: 404
    data: '{"error": "User not found"}'
```

**[📖 Scenarios Documentation](docs/scenarios.md)**

## Installation Methods

| Method | Command | Use Case |
|--------|---------|----------|
| **Docker** | `docker run -p 8080:8080 ghcr.io/bmcszk/unimock` | Quick testing |
| **Kubernetes** | `helm install unimock ./helm/unimock` | Production testing |
| **Local Development** | `make tilt-run` | Development with auto-reload |
| **Go Library** | `import "github.com/bmcszk/unimock/pkg"` | Embed in Go applications |

**[📖 Deployment Guide](docs/deployment.md)**

## Using Unimock

### As a Go Library

Embed Unimock directly in your Go application:

```go
import (
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
)

// Create configuration
mockConfig := &config.MockConfig{
    Sections: map[string]config.Section{
        "users": {
            PathPattern: "/api/users/*",
            BodyIDPaths: []string{"/id"},  // XPath-like syntax, not JSONPath
            ReturnBody:  true,
        },
    },
}

// Start embedded server
server, err := pkg.NewServer(
    pkg.WithPort(8080),
    pkg.WithMockConfig(mockConfig),
)
if err != nil {
    log.Fatal(err)
}

go server.ListenAndServe()
```

### With Client Libraries

Use Unimock from your tests:

```go
// Go
client := unimock.NewClient("http://localhost:8080")
user := client.Post("/api/users", userData)
```

**[📖 Full Client Guide](docs/client.md)**

## Advanced Features

- **[Data Transformations](docs/transformations.md)** - Modify responses on-the-fly
- **[Request Matching](docs/matching.md)** - Advanced path and content matching
- **[Monitoring](docs/monitoring.md)** - Metrics and health checks

## API Reference

### Technical Endpoints
- `GET /_uni/health` - Health check
- `GET /_uni/metrics` - Prometheus metrics
- `GET /_uni/scenarios` - List active scenarios

### Mock Operations
- `POST /your/path` - Add mock data
- `GET /your/path/{id}` - Retrieve mock data
- `PUT /your/path/{id}` - Update mock data
- `DELETE /your/path/{id}` - Delete mock data

## Development

```bash
# Clone and build
git clone https://github.com/bmcszk/unimock.git
cd unimock
make build

# Run tests
make test

# Start with auto-reload
make tilt-run
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run `make check` to validate changes
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.