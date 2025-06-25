# Deployment Guide

This guide covers different ways to deploy and run Unimock.

## Quick Start Options

| Method | Best For | Command |
|--------|----------|---------|
| Docker | Quick testing | `docker run -p 8080:8080 ghcr.io/bmcszk/unimock` |
| Go Binary | Go developers | `go run .` |
| Kubernetes | Production testing | `helm install unimock ./helm/unimock` |
| Tilt | Local development | `make tilt-run` |

## Docker Deployment

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

### Available Tags

- `latest` - Latest stable release
- `v1.x.x` - Specific version tags
- `main` - Latest development build

## Kubernetes with Helm

### Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

### Basic Installation

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
  sections:
    - path: "/api/users/*"
      id_path: "$.id"
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
kubectl port-forward svc/my-unimock 8080:8080

# Test
curl http://localhost:8080/_uni/health
```

### Uninstall

```bash
helm uninstall my-unimock
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

## Go Binary

### Install from Source

```bash
# Clone and build
git clone https://github.com/bmcszk/unimock.git
cd unimock
make build

# Run
./unimock
```

### Install with Go

```bash
go install github.com/bmcszk/unimock@latest
unimock
```

### Configuration

By default, looks for `config.yaml` in current directory:

```bash
# Use custom config
UNIMOCK_CONFIG=/path/to/config.yaml unimock

# Use different port
UNIMOCK_PORT=9090 unimock

# Enable debug logging
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
  sections:
    - path: "/api/v1/users/*"
      id_path: "$.id"
      return_body: true
      strict_path: false

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
helm install unimock-prod ./helm/unimock -f production-values.yaml
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

### Grafana Dashboard

Import the provided dashboard from `monitoring/grafana-dashboard.json` for visualizing Unimock metrics.

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