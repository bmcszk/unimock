# Unimock Helm Chart

This Helm chart deploys the Unimock HTTP mock server on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Getting Started

### Adding the Chart

```bash
# Clone the repository
git clone https://github.com/bmcszk/unimock.git
cd unimock

# Install the chart
helm install my-unimock ./helm/unimock
```

### Uninstalling the Chart

```bash
helm uninstall my-unimock
```

## Configuration

The following table lists the configurable parameters of the Unimock chart and their default values.

| Parameter                 | Description                                   | Default                            |
|---------------------------|-----------------------------------------------|-----------------------------------|
| `replicaCount`            | Number of replicas                            | `1`                               |
| `image.repository`        | Image repository                              | `unimock`                         |
| `image.tag`               | Image tag                                     | `latest`                          |
| `image.pullPolicy`        | Image pull policy                             | `IfNotPresent`                    |
| `service.type`            | Kubernetes Service type                       | `ClusterIP`                       |
| `service.port`            | Service port                                  | `8080`                            |
| `ingress.enabled`         | Enable ingress                                | `false`                           |
| `env.UNIMOCK_PORT`        | Port to listen on                             | `"8080"`                          |
| `env.UNIMOCK_LOG_LEVEL`   | Log level                                     | `"info"`                          |
| `config.yaml`             | Unimock configuration                         | Default mock config (see values.yaml) |

## Usage

### Accessing the Service

If using the default configuration with `ClusterIP` service type:

```bash
# Port forward to access locally
kubectl port-forward svc/my-unimock-unimock 8080:8080

# Test the service
curl http://localhost:8080/api/users
```

If Ingress is enabled, you can access the service at the configured hostname.

### Using Tech Endpoints

Unimock provides technical endpoints for monitoring:

```bash
# Health check
curl http://localhost:8080/_uni/health

# Metrics
curl http://localhost:8080/_uni/metrics
```

### Adding Mock Data

```bash
# Add a user
curl -X POST -H "Content-Type: application/json" \
  -d '{"id": "1", "name": "John Doe", "email": "john@example.com"}' \
  http://localhost:8080/api/users

# Retrieve user
curl http://localhost:8080/api/users/1
``` 
