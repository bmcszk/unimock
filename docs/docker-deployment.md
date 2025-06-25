# Docker Deployment Guide

This guide covers deploying Unimock using Docker and GitHub Container Registry.

## Automatic Builds

Unimock automatically builds and publishes Docker images to GitHub Container Registry (ghcr.io) via GitHub Actions:

- **On version tags**: Creates versioned tags (e.g., `v1.2.0`, `v1.2`, `v1`) and updates `latest`
- **On pull requests**: Builds images for testing (not published)

## Available Images

All images are available at: `ghcr.io/bmcszk/unimock`

### Tags

- `latest` - Latest stable release (from latest version tag)
- `v1.x.x` - Specific version releases (e.g., `v1.2.0`)
- `v1.x` - Minor version tags (e.g., `v1.2`)
- `v1` - Major version tags (e.g., `v1`)

### Platforms

Images are built for multiple architectures:
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, including Apple Silicon)

## Deployment Examples

### Basic Deployment

```bash
# Run with default configuration
docker run -p 8080:8080 ghcr.io/bmcszk/unimock:latest
```

### Production Deployment with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  unimock:
    image: ghcr.io/bmcszk/unimock:latest
    container_name: unimock
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config/config.yaml:/etc/unimock/config.yaml:ro
      - ./config/scenarios.yaml:/etc/unimock/scenarios.yaml:ro
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
    networks:
      - unimock-network

  # Optional: Add nginx reverse proxy
  nginx:
    image: nginx:alpine
    container_name: unimock-proxy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - unimock
    networks:
      - unimock-network

networks:
  unimock-network:
    driver: bridge
```

### Kubernetes Deployment

```yaml
# kubernetes-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: unimock
  labels:
    app: unimock
spec:
  replicas: 2
  selector:
    matchLabels:
      app: unimock
  template:
    metadata:
      labels:
        app: unimock
    spec:
      containers:
      - name: unimock
        image: ghcr.io/bmcszk/unimock:latest
        ports:
        - containerPort: 8080
        env:
        - name: UNIMOCK_PORT
          value: "8080"
        - name: UNIMOCK_CONFIG
          value: "/etc/unimock/config.yaml"
        - name: UNIMOCK_LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: config
          mountPath: /etc/unimock/config.yaml
          subPath: config.yaml
          readOnly: true
        - name: scenarios
          mountPath: /etc/unimock/scenarios.yaml
          subPath: scenarios.yaml
          readOnly: true
        livenessProbe:
          httpGet:
            path: /_uni/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /_uni/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: unimock-config
      - name: scenarios
        configMap:
          name: unimock-scenarios

---
apiVersion: v1
kind: Service
metadata:
  name: unimock-service
spec:
  selector:
    app: unimock
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: unimock-config
data:
  config.yaml: |
    sections:
      - path: "/api/users/*"
        id_path: "/id"
      - path: "/api/products/*"
        id_path: "/id"

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: unimock-scenarios
data:
  scenarios.yaml: |
    scenarios:
      - uuid: "health-check"
        method: "GET"
        path: "/_uni/health"
        status_code: 200
        content_type: "application/json"
        data: '{"status":"ok"}'
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `UNIMOCK_PORT` | Server port | `8080` |
| `UNIMOCK_CONFIG` | Path to config file | `/etc/unimock/config.yaml` |
| `UNIMOCK_SCENARIOS_FILE` | Path to scenarios file | (optional) |
| `UNIMOCK_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

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

## Monitoring and Observability

### Health Checks

Unimock provides health check endpoints:

```bash
# Basic health check
curl http://localhost:8080/_uni/health

# Metrics endpoint
curl http://localhost:8080/_uni/metrics
```

### Logging

Configure structured logging:

```yaml
environment:
  - UNIMOCK_LOG_LEVEL=info  # debug, info, warn, error
```

### Prometheus Metrics

The `/_uni/metrics` endpoint provides metrics in a format suitable for Prometheus scraping.

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   # Ensure config files are readable
   chmod 644 config.yaml scenarios.yaml
   ```

2. **Port Already in Use**
   ```bash
   # Check what's using port 8080
   lsof -i :8080
   
   # Use different port
   docker run -p 9090:8080 -e UNIMOCK_PORT=8080 ghcr.io/bmcszk/unimock:latest
   ```

3. **Config File Not Found**
   ```bash
   # Verify volume mounting
   docker run -v $(pwd)/config.yaml:/etc/unimock/config.yaml ghcr.io/bmcszk/unimock:latest
   ```

### Debug Mode

```bash
# Run with debug logging
docker run -p 8080:8080 \
  -e UNIMOCK_LOG_LEVEL=debug \
  ghcr.io/bmcszk/unimock:latest
```

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