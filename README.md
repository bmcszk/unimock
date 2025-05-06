# Unimock

A Universal Mock HTTP server designed for e2e testing scenarios where you need to mock third-party REST services. This server can work as a RESTful service, XML service, or any other HTTP-based service.

## Purpose

Unimock was created to solve a common problem in e2e testing: the need to mock third-party REST services that you don't control. Instead of building custom mock services for each test scenario, Unimock provides a universal solution that can handle any HTTP-based service.

## Key Features

- **Universal Support**: Works with any HTTP-based service (REST, XML, etc.)
- **Flexible ID Extraction**: Extracts IDs from headers, JSON/XML bodies, or paths
- **Thread-Safe Storage**: In-memory storage with mutex protection
- **Full HTTP Support**: All HTTP methods (GET, POST, PUT, DELETE)
- **Technical Endpoints**: Health check, metrics, and scenario management
- **Scenario-Based Mocking**: Create predefined scenarios for paths that bypass regular mock behavior

## Installation

```bash
go get github.com/bmcszk/unimock
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
