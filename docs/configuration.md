# Configuration

Unimock is configured using a YAML configuration file. This document describes the configuration options and format.

## Configuration File Format

The configuration file uses YAML format and contains sections for different API endpoint patterns.

Example:

```yaml
sections:
  users:
    path_pattern: "/users/*"
    header_id_name: "X-Resource-ID"
    body_id_paths:
      - "/id"
      - "/data/id"
      - "//id"
      - "/items/*/id"
  
  orders:
    path_pattern: "/orders/*"
    header_id_name: "X-Order-ID"
    body_id_paths:
      - "/order/id"
      - "//id"
```

## Section Configuration

Each section has the following properties:

### Required Properties

- `path_pattern` - The URL pattern to match (e.g., `/users/*`, `/users/*/orders/*`)

### Optional Properties

- `header_id_name` - The HTTP header name to extract ID from (e.g., `X-Resource-ID`)
- `body_id_paths` - Array of XPath-like paths to extract IDs from request body

## Environment Variables

Unimock can be configured with the following environment variables:

- `UNIMOCK_PORT` - The port to listen on (default: `8080`)
- `UNIMOCK_CONFIG` - The path to the configuration file (default: `config.yaml`)
- `UNIMOCK_LOG_LEVEL` - The log level (default: `info`)

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
header_id_name: "X-Resource-ID"
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
