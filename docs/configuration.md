# Configuration

Unimock is configured using a YAML configuration file. This document describes the configuration options and format.

## Configuration File Format

Unimock supports two configuration formats:

### 1. Unified Format (Recommended)

The unified format allows you to define both sections and scenarios in a single file:

```yaml
sections:
  users:
    path_pattern: "/api/users/*"
    header_id_names: ["X-User-ID", "Authorization"] # Multiple headers supported
    body_id_paths:
      - "/id"
      - "/user/id"
      - "/@id"
      - "/data/id"
    return_body: true
  
  orders:
    path_pattern: "/api/orders/*"
    header_id_names: ["X-Order-ID"]
    body_id_paths:
      - "/id"
      - "/order/id"

scenarios:
  - uuid: "user-not-found"
    method: "GET"
    path: "/api/users/999"
    response:
      status_code: 404
      body: '{"error": "User not found"}'
```

### 2. Legacy Format (Still Supported)

The legacy format defines sections at the root level:

```yaml
users:
  path_pattern: "/api/users/*"
  header_id_names: ["X-User-ID"]
  body_id_paths:
    - "/id"
    - "/user/id"

orders:
  path_pattern: "/api/orders/*"
  header_id_names: ["X-Order-ID"]
  body_id_paths:
    - "/id"
    - "/order/id"
```

## Section Configuration

Each section has the following properties:

### Required Properties

- `path_pattern` - The URL pattern to match (e.g., `/api/users/*`, `/api/users/*/orders/*`)

### Optional Properties

- `header_id_names` - Array of HTTP header names to extract IDs from (e.g., `["X-User-ID", "Authorization"]`)
- `body_id_paths` - Array of XPath-like paths to extract IDs from request body (e.g., `["/id", "/user/id", "/@id"]`)
- `return_body` - Whether to return the request body in responses (default: false)

### ID Extraction

IDs can be extracted from multiple sources:

1. **URL Path**: Automatically extracted from wildcards in `path_pattern`
2. **HTTP Headers**: From any headers listed in `header_id_names`
3. **Request Body**: From JSON/XML paths specified in `body_id_paths`

Path syntax supports:
- `/id` - Extract from JSON field `id`
- `/@id` - Extract from XML attribute `id`
- `/user/id` - Extract from nested JSON field `user.id`
- `/items/*/id` - Extract from array elements

## Environment Variables

Unimock can be configured with the following environment variables:

- `UNIMOCK_PORT` - The port to listen on (default: `8080`)
- `UNIMOCK_CONFIG` - The path to the configuration file (default: `config.yaml`)
- `UNIMOCK_LOG_LEVEL` - The log level: `debug`, `info`, `warn`, `error` (default: `info`)

## Scenarios

Scenarios allow you to define fixed responses for specific requests, useful for testing edge cases or error conditions. They can be defined in the unified configuration format or loaded at runtime via the API.

See [Scenarios Guide](scenarios.md) for detailed information.

## Path Patterns

Path patterns use wildcard notation to match API endpoints:

- `*` - Matches any segment (used for ID segments)
- Examples:
  - `/users/*` - Matches `/users/123`
  - `/users/*/orders/*` - Matches `/users/123/orders/456`

## ID Extraction Configuration

### Header-based ID

Specify the HTTP header name to extract IDs from:

```yaml
header_id_names: ["X-Resource-ID"]
```

### Body-based ID

Specify XPath-like paths to extract IDs from request bodies (works for both JSON and XML):

```yaml
body_id_paths:
  - "/id"           # Root level ID
  - "/data/id"      # Nested ID
  - "//id"          # Any ID anywhere
  - "/items/*/id"   # Array of objects with IDs
```

## Configuration Loading

1. Unimock looks for the configuration file at startup
2. The file path can be specified with the `UNIMOCK_CONFIG` environment variable
3. If not specified, it defaults to `config.yaml` in the current directory
4. If the configuration file is invalid or missing, Unimock will log an error and exit

## Example Configurations

### Basic REST API

```yaml
sections:
  users:
    path_pattern: "/users/*"
    body_id_paths:
      - "/id"
  
  orders:
    path_pattern: "/orders/*"
    body_id_paths:
      - "/id"
```

### XML-based Service

```yaml
sections:
  products:
    path_pattern: "/products/*"
    body_id_paths:
      - "/product/id"
      - "//id"
```

### Complex Nested IDs

```yaml
sections:
  organization:
    path_pattern: "/orgs/*"
    body_id_paths:
      - "/organization/id"
      - "/data/organization/id"
      - "//id"
``` 
