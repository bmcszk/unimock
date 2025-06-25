# Unimock Helm Chart

A Helm chart for deploying Unimock - Universal HTTP mock server for e2e testing scenarios.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Installing the Chart

To install the chart with the release name `my-unimock`:

```bash
# Add the chart repository (if published)
helm repo add unimock https://charts.unimock.dev
helm repo update

# Or install directly from source
git clone https://github.com/bmcszk/unimock.git
cd unimock/helm

# Install with default values
helm install my-unimock ./unimock

# Install with custom values
helm install my-unimock ./unimock -f values.yaml

# Install with inline values
helm install my-unimock ./unimock \
  --set image.tag=v1.2.0 \
  --set scenarios.enabled=true
```

## Uninstalling the Chart

To uninstall/delete the `my-unimock` deployment:

```bash
helm uninstall my-unimock
```

## Configuration

The following table lists the configurable parameters of the Unimock chart and their default values.

### Image Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Unimock image repository | `ghcr.io/bmcszk/unimock` |
| `image.tag` | Unimock image tag | `""` (uses appVersion) |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets | `[]` |

### Deployment Configuration

**Note**: Unimock runs as single instance only (hardcoded to 1 replica).

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nameOverride` | Override name | `""` |
| `fullnameOverride` | Override full name | `""` |

### Service Account

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.automount` | Automount service account token | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name | `""` |

### Security Context

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.fsGroup` | Pod filesystem group ID | `65534` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.capabilities.drop` | Dropped capabilities | `["ALL"]` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `true` |
| `securityContext.runAsNonRoot` | Run as non-root user | `true` |
| `securityContext.runAsUser` | User ID | `65534` |
| `securityContext.runAsGroup` | Group ID | `65534` |

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `service.targetPort` | Target port | `http` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts | `[{host: "unimock.local", paths: [{path: "/", pathType: "Prefix"}]}]` |
| `ingress.tls` | Ingress TLS configuration | `[]` |

### Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `256Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |

### Note on Scaling

**Important**: Unimock is designed for testing scenarios and only supports single replica deployment. Autoscaling and Pod Disruption Budgets are not applicable.

### Environment Variables

| Parameter | Description | Default |
|-----------|-------------|---------|
| `env.UNIMOCK_PORT` | Server port | `"8080"` |
| `env.UNIMOCK_LOG_LEVEL` | Log level | `"info"` |

### Unimock Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.sections` | Mock configuration sections (YAML object) | See values.yaml |
| `config.yaml` | Raw configuration YAML string | `""` |

### Scenarios Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `scenarios.enabled` | Enable predefined scenarios | `false` |
| `scenarios.data` | Scenarios configuration (YAML object) | See values.yaml |
| `scenarios.yaml` | Raw scenarios YAML string | `""` |

### Health Probes

| Parameter | Description | Default |
|-----------|-------------|---------|
| `probes.liveness.enabled` | Enable liveness probe | `true` |
| `probes.readiness.enabled` | Enable readiness probe | `true` |
| `probes.startup.enabled` | Enable startup probe | `true` |

### Monitoring

| Parameter | Description | Default |
|-----------|-------------|---------|
| `monitoring.serviceMonitor.enabled` | Enable ServiceMonitor for Prometheus | `false` |
| `monitoring.serviceMonitor.namespace` | ServiceMonitor namespace | `""` |
| `monitoring.serviceMonitor.interval` | Scrape interval | `30s` |

## Examples

### Basic Deployment

```yaml
# values.yaml

config:
  sections:
    - path: "/api/users/*"
      id_path: "$.id"
      return_body: true
    - path: "/api/products/*"
      id_path: "$.product_id"
      return_body: true

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### With Predefined Scenarios

```yaml
# values.yaml
scenarios:
  enabled: true
  data:
    scenarios:
      - uuid: "health-check"
        method: "GET"
        path: "/_uni/health"
        status_code: 200
        content_type: "application/json"
        data: '{"status":"ok"}'
      - uuid: "sample-user"
        method: "GET"
        path: "/api/users/sample"
        status_code: 200
        content_type: "application/json"
        data: |
          {
            "id": "sample",
            "name": "Sample User",
            "email": "sample@example.com"
          }
```

### With Ingress and TLS

```yaml
# values.yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: unimock.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: unimock-tls
      hosts:
        - unimock.example.com
```

### Production Configuration

```yaml
# values-prod.yaml

image:
  tag: "v1.2.0"

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi

# Autoscaling not supported for single replica testing
# autoscaling:
#   enabled: false

monitoring:
  serviceMonitor:
    enabled: true
    namespace: monitoring

config:
  sections:
    - path: "/api/v1/users/*"
      id_path: "$.id"
      return_body: true
      strict_path: false
    - path: "/api/v1/orders/*"
      id_path: "$.order_id"
      header_id_name: "X-Order-ID"
      return_body: true
      strict_path: false

scenarios:
  enabled: true
  data:
    scenarios:
      - uuid: "health-check"
        method: "GET"
        path: "/_uni/health"
        status_code: 200
        content_type: "application/json"
        data: '{"status":"ok","environment":"production"}'
```

## Configuration Formats

### YAML Object Format (Recommended)

Use structured YAML objects for better validation and IDE support:

```yaml
config:
  sections:
    - path: "/api/users/*"
      id_path: "$.id"
      return_body: true

scenarios:
  enabled: true
  data:
    scenarios:
      - uuid: "test-scenario"
        method: "GET"
        path: "/test"
        status_code: 200
        data: "test response"
```

### Raw YAML String Format

Alternatively, use raw YAML strings:

```yaml
config:
  yaml: |
    sections:
      - path: "/api/users/*"
        id_path: "$.id"

scenarios:
  enabled: true
  yaml: |
    scenarios:
      - uuid: "test-scenario"
        method: "GET"
        path: "/test"
        status_code: 200
        data: "test response"
```

## Troubleshooting

### Common Issues

1. **Pod won't start**: Check resource limits and pull secrets
   ```bash
   kubectl describe pod -l app.kubernetes.io/name=unimock
   ```

2. **Config not loading**: Verify ConfigMap content
   ```bash
   kubectl get configmap my-unimock-config -o yaml
   ```

3. **Health check failing**: Check probe configuration and app startup time
   ```bash
   kubectl logs -l app.kubernetes.io/name=unimock
   ```

### Debug Commands

```bash
# Check deployment status
kubectl get deployment my-unimock

# View pod logs
kubectl logs -l app.kubernetes.io/name=unimock -f

# Check service endpoints
kubectl get endpoints my-unimock

# Test health endpoint
kubectl port-forward svc/my-unimock 8080:8080 &
curl http://localhost:8080/_uni/health

# View configuration
kubectl get configmap my-unimock-config -o yaml
kubectl get configmap my-unimock-scenarios -o yaml
```

## Chart Development

### Linting and Testing

```bash
# Lint the chart
helm lint ./helm/unimock

# Test template rendering
helm template my-unimock ./helm/unimock

# Test with different values
helm template my-unimock ./helm/unimock -f test-values.yaml

# Dry run install
helm install my-unimock ./helm/unimock --dry-run --debug
```

### Packaging

```bash
# Package the chart
helm package ./helm/unimock

# Check chart content
helm show all ./unimock-1.0.0.tgz
```