# Deployment Guide

This guide covers different ways to deploy and run Unimock.

## Quick Start Options

| Method | Best For | Command |
|--------|----------|---------|
| **Docker** | Quick testing | `docker run -p 8080:8080 ghcr.io/bmcszk/unimock` |
| **Helm (Registry)** | Production deployment | `helm install unimock oci://ghcr.io/bmcszk/charts/unimock` |
| **Helm (Local)** | Local Kubernetes testing | `helm install unimock ./helm/unimock` |
| **Go Library** | Embedded in Go apps | `pkg.NewServer()` |
| **Tilt** | Local development | `make tilt-run` |

## Docker Deployment

### Available Images

All images are available at: `ghcr.io/bmcszk/unimock`

**Tags:**
- `latest` - Latest stable release (from latest version tag)
- `v1.x.x` - Specific version releases (e.g., `v1.2.0`)
- `v1.x` - Minor version tags (e.g., `v1.2`)
- `v1` - Major version tags (e.g., `v1`)

**Platforms:**
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, including Apple Silicon)

### Basic Usage

```bash
# Run with defaults
docker run -p 8080:8080 ghcr.io/bmcszk/unimock:latest

# Run with custom config
docker run -p 8080:8080 \
  -v $(pwd)/my-config.yaml:/etc/unimock/config.yaml \
  ghcr.io/bmcszk/unimock:latest

# Run with scenarios
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/unimock/config.yaml \
  -v $(pwd)/scenarios.yaml:/etc/unimock/scenarios.yaml \
  -e UNIMOCK_SCENARIOS_FILE=/etc/unimock/scenarios.yaml \
  ghcr.io/bmcszk/unimock:latest
```

### Docker Compose

Create `docker-compose.yml`:

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
      - UNIMOCK_LOG_LEVEL=debug
      - UNIMOCK_SCENARIOS_FILE=/etc/unimock/scenarios.yaml
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/_uni/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

Start with:
```bash
docker-compose up -d
```

## Kubernetes with Helm

### Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

### Installation from Registry (Recommended)

```bash
# Install latest version from GitHub Container Registry
helm install unimock oci://ghcr.io/bmcszk/charts/unimock

# Install specific version
helm install unimock oci://ghcr.io/bmcszk/charts/unimock --version 1.2.0

# Install with custom values
helm install unimock oci://ghcr.io/bmcszk/charts/unimock -f my-values.yaml
```

### Installation from Local Chart

```bash
# Clone repository
git clone https://github.com/bmcszk/unimock.git
cd unimock

# Install with defaults
helm install my-unimock ./helm/unimock

# Install with custom values
helm install my-unimock ./helm/unimock -f my-values.yaml
```

### Custom Values Example

Create `my-values.yaml`:

```yaml
# Use specific version
image:
  tag: "v1.2.0"

# Enable scenarios
scenarios:
  enabled: true
  data:
    scenarios:
      - uuid: "health-check"
        method: "GET"
        path: "/_uni/health"
        status_code: 200
        data: '{"status": "ok"}'

# Custom configuration
config:
  users:
    path_pattern: "/api/users/*"
    body_id_paths:
      - "/id"
    return_body: true

# Enable ingress
ingress:
  enabled: true
  hosts:
    - host: unimock.mycompany.com
      paths:
        - path: /
          pathType: Prefix

# Resource limits
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

### Access the Service

```bash
# Port forward for local access
kubectl port-forward svc/unimock 8080:8080

# Test
curl http://localhost:8080/_uni/health
```

### Uninstall

```bash
helm uninstall unimock
```

## Local Development with Tilt

Tilt provides live-reload development with Kubernetes.

### Prerequisites

- [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker)
- [Tilt](https://tilt.dev/)
- [Helm](https://helm.sh/)

### Quick Start

```bash
# Start development environment
make tilt-run

# This will:
# 1. Create Kind cluster
# 2. Build Docker image
# 3. Deploy to Kubernetes
# 4. Set up port forwarding
# 5. Enable live reload
```

### Access Tilt UI

Open http://localhost:10350 to see:
- Build status
- Pod logs
- Resource status
- Manual commands

### Stop Development

```bash
make tilt-stop
```

### CI Mode

For automated testing:

```bash
make tilt-ci
```

## Go Library

Embed Unimock directly in your Go application for testing, development, or production use. The library provides full programmatic control including data transformations, scenario management, and embedded server capabilities.

**[📖 Complete Go Library Guide](library.md)**

### Quick Example

```go
import "github.com/bmcszk/unimock/pkg"

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

### Go Binary

For standalone usage:

```bash
# Install from source
git clone https://github.com/bmcszk/unimock.git
cd unimock
make build
./unimock

# Install with Go
go install github.com/bmcszk/unimock@latest
unimock

# Configuration via environment variables
UNIMOCK_CONFIG=/path/to/config.yaml unimock
UNIMOCK_PORT=9090 unimock
UNIMOCK_LOG_LEVEL=debug unimock
```

## Production Deployment

### Kubernetes Production Setup

```yaml
# production-values.yaml
image:
  tag: "v1.2.0"  # Pin to specific version

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 512Mi

# Enable monitoring
monitoring:
  serviceMonitor:
    enabled: true
    namespace: monitoring

# Production config
config:
  users:
    path_pattern: "/api/v1/users/*"
    body_id_paths:
      - "/id"
    return_body: true

# Health checks
probes:
  liveness:
    initialDelaySeconds: 30
    periodSeconds: 10
  readiness:
    initialDelaySeconds: 5
    periodSeconds: 5

# Ingress with TLS
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: unimock.prod.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: unimock-tls
      hosts:
        - unimock.prod.example.com
```

Deploy:
```bash
helm install unimock-prod oci://ghcr.io/bmcszk/charts/unimock -f production-values.yaml
```

### Docker Swarm

```yaml
# docker-stack.yml
version: '3.8'
services:
  unimock:
    image: ghcr.io/bmcszk/unimock:v1.2.0
    ports:
      - "8080:8080"
    configs:
      - source: unimock-config
        target: /etc/unimock/config.yaml
    environment:
      - UNIMOCK_LOG_LEVEL=info
    deploy:
      replicas: 1  # Unimock doesn't support clustering
      resources:
        limits:
          memory: 256M
        reservations:
          memory: 128M
      restart_policy:
        condition: on-failure

configs:
  unimock-config:
    file: ./config.yaml
```

Deploy:
```bash
docker stack deploy -c docker-stack.yml unimock
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `UNIMOCK_PORT` | Server port | `8080` |
| `UNIMOCK_CONFIG` | Config file path | `config.yaml` |
| `UNIMOCK_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `UNIMOCK_SCENARIOS_FILE` | Scenarios file path | (disabled) |

## Security Considerations

### Image Security

- Uses distroless base image for minimal attack surface
- Runs as non-root user
- Images are scanned with Trivy for vulnerabilities
- Static binary with no dynamic dependencies

### Runtime Security

```yaml
# Example security-hardened container
services:
  unimock:
    image: ghcr.io/bmcszk/unimock:latest
    read_only: true
    cap_drop:
      - ALL
    security_opt:
      - no-new-privileges:true
    user: "65534:65534"  # nobody user
    tmpfs:
      - /tmp:rw,noexec,nosuid,size=100m
```

## Health Checks

All deployment methods support health checking:

```bash
# Health endpoint
curl http://localhost:8080/_uni/health

# Returns:
# {"status": "ok", "service": "unimock"}
```

Use this endpoint for:
- Docker healthchecks
- Kubernetes probes
- Load balancer health checks

## Monitoring

### Prometheus Metrics

```bash
curl http://localhost:8080/_uni/metrics
```

### Common Metrics

- `unimock_requests_total` - Total HTTP requests
- `unimock_request_duration_seconds` - Request latency
- `unimock_active_mocks` - Number of stored mocks

## Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check logs
docker logs <container-id>
kubectl logs -l app.kubernetes.io/name=unimock
```

**Config not loading:**
```bash
# Verify file paths
ls -la /etc/unimock/
# Check logs for config errors
UNIMOCK_LOG_LEVEL=debug unimock
```

**Can't reach service:**
```bash
# Check if service is running
curl http://localhost:8080/_uni/health

# Check port forwarding
kubectl get svc
kubectl port-forward svc/unimock 8080:8080
```

### Debug Mode

Enable debug logging for detailed information:

```bash
UNIMOCK_LOG_LEVEL=debug unimock
```

This shows:
- Config loading details
- Request routing decisions
- ID extraction attempts
- Response selection logic

## Building from Source

### Local Build

```bash
# Clone repository
git clone https://github.com/bmcszk/unimock.git
cd unimock

# Build image
docker build -t unimock:local .

# Run locally built image
docker run -p 8080:8080 unimock:local
```

### Multi-platform Build

```bash
# Setup buildx for multi-platform builds
docker buildx create --use

# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 \
  -t unimock:multi-platform .
```