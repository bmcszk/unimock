# Unimock - Universal HTTP Mock Server

[![Pull request tests](https://github.com/bmcszk/unimock/actions/workflows/pr.yml/badge.svg)](https://github.com/bmcszk/unimock/actions/workflows/pr.yml)
[![Push tests](https://github.com/bmcszk/unimock/actions/workflows/push.yml/badge.svg)](https://github.com/bmcszk/unimock/actions/workflows/push.yml)
[![Security](https://github.com/bmcszk/unimock/actions/workflows/security.yml/badge.svg?branch=master)](https://github.com/bmcszk/unimock/actions/workflows/security.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/bmcszk/unimock/badge)](https://scorecard.dev/viewer/?uri=github.com/bmcszk/unimock)
[![coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/bmcszk/unimock/gh-pages/badges/coverage.json&cacheSeconds=3600)](https://github.com/bmcszk/unimock/blob/master/.github/workflows/push.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmcszk/unimock)](https://goreportcard.com/report/github.com/bmcszk/unimock)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmcszk/unimock.svg)](https://pkg.go.dev/github.com/bmcszk/unimock/client)
[![Release](https://img.shields.io/github/v/release/bmcszk/unimock)](https://github.com/bmcszk/unimock/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26%20%7C%201.27-00ADD8?logo=go)](go.mod)

**Mock any HTTP service for testing. Works with REST, GraphQL, XML, or any HTTP API.**

## Quick Start

```bash
# 1. Run with Docker
docker run -p 8080:8080 ghcr.io/bmcszk/unimock:latest

# 2. Test it
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "John Doe"}'

curl http://localhost:8080/api/users/123
```

## Configuration

Create `config.yaml` with the unified format:

```yaml
sections:
  users:
    path_pattern: "/api/users/*"
    body_id_paths: ["/id"]
    header_id_names: ["X-User-ID", "Authorization"]  # Multiple headers supported
    return_body: true

  orders:
    path_pattern: "/api/orders/*"
    body_id_paths: ["/order_id"]
    header_id_names: ["X-Order-ID"]
```

**➜ [Copy-paste examples](examples/configs/) | [Full config guide](docs/configuration.md)**

## Scenarios

Pre-define responses for testing:

```yaml
sections:
  users:
    path_pattern: "/api/users/*"
    body_id_paths: ["/id"]

scenarios:
  - uuid: "user-not-found"
    method: "GET"
    path: "/api/users/999"
    status_code: 404
    content_type: "application/json"
    data: '{"error": "User not found"}'
```

**➜ [Scenario examples](docs/scenarios.md)**

## Installation

| Method | Command | Use Case |
|--------|---------|----------|
| **Docker** | `docker run -p 8080:8080 ghcr.io/bmcszk/unimock` | Quick testing |
| **Helm (Registry)** | `helm install unimock oci://ghcr.io/bmcszk/charts/unimock` | Production |  
| **Local Build** | `make build && ./unimock` | Development |
| **Go Library** | `import "github.com/bmcszk/unimock/pkg"` | Embed in apps |

## Key Features

✅ **Universal** - Works with any HTTP service (REST, XML, GraphQL)  
✅ **Smart ID extraction** - From URLs, JSON, XML, or multiple headers  
✅ **Scenarios** - Fixed responses for testing edge cases  
✅ **Thread-safe** - Handles concurrent requests  
✅ **Simple config** - Single YAML file  

## Quick Reference

| Task | Command/Config |
|------|---------------|
| **Run** | `docker run -p 8080:8080 ghcr.io/bmcszk/unimock` |
| **Configure** | Edit `config.yaml` with sections and scenarios |
| **Health** | `GET /_uni/health` |
| **Metrics** | `GET /_uni/metrics` |
| **POST test** | `curl -X POST :8080/api/users -d '{"id":"1"}'` |
| **GET test** | `curl :8080/api/users/1` |

## Examples

- **[Simple Examples](examples/configs/config-simple.yaml)** - Copy-paste configurations
- **[Unified Format](examples/configs/config-unified.yaml)** - Sections + scenarios in one file
- **[Library Usage](examples/library/)** - Go library integration example
- **[Legacy Format](config.yaml)** - Traditional format (still supported)

## Environment Variables

- `UNIMOCK_PORT` - Server port (default: 8080)
- `UNIMOCK_LOG_LEVEL` - Log level (default: info)  
- `UNIMOCK_CONFIG` - Config file path (default: config.yaml)

## Common Use Cases

- **API Testing** - Mock external services in integration tests
- **Development** - Work offline without real API dependencies  
- **Error Testing** - Simulate API failures and edge cases
- **Load Testing** - Fast, predictable responses

## Documentation

📖 **[Configuration Guide](docs/configuration.md)** - Complete config reference  
📖 **[Scenarios Guide](docs/scenarios.md)** - Pre-defined responses  
📖 **[Examples](docs/examples.md)** - Usage examples  
📖 **[Go Library](docs/library.md)** - Embed in applications  
📖 **[Deployment](docs/deployment.md)** - Production deployment  

## Development

```bash
git clone https://github.com/bmcszk/unimock.git
cd unimock
make build      # Build binary
make test       # Run tests  
make check      # Full validation
```

## License

MIT License - see [LICENSE](LICENSE)