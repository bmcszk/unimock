# ID Extraction

This document explains how Unimock extracts IDs from HTTP requests for resource identification and storage.

## Overview

Unimock can extract IDs from three sources:
1. **URL path** - From the path segment (e.g., `/users/123`)
2. **HTTP headers** - From named headers (e.g., `X-User-ID: 123`)
3. **Request body** - From JSON/XML content using XPath-like expressions

The extraction is configured per API endpoint pattern using sections in the main configuration file.

## Configuration

Each section in `config.yaml` defines how to extract IDs for a specific API endpoint pattern:

```yaml
# Example configuration showing all ID extraction methods
sections:
  users:
    path_pattern: "/api/users/*"  # URL pattern to match
    header_id_names: ["X-User-ID"]   # Optional: HTTP headers to extract ID from
    body_id_paths:                # Optional: XPath-like paths to extract IDs from request body
      - "/id"                     # Root level ID
      - "/user/id"                # Nested ID
      - "//id"                    # Any ID anywhere in document
    return_body: true             # Optional: return POST body on GET requests
```

### Path Pattern

The `path_pattern` defines which URLs this section applies to:
- Use `*` as a wildcard for ID segments
- Examples:
  - `/users/*` - matches `/users/123`
  - `/users/*/orders/*` - matches `/users/123/orders/456`

### Header ID Extraction

If `header_id_names` is specified, Unimock will try to extract IDs from the named HTTP headers:
- Example: `X-Resource-ID: 123` will extract `123` as the ID
- If the header is not present or empty, this extraction method is skipped

### Body ID Extraction

The `body_id_paths` array defines XPath-like paths to extract IDs from the request body. The paths work for both JSON and XML content types.

#### JSON Path Syntax

For JSON requests, use XPath-like syntax:
- `/` - Start from root
- `//` - Search anywhere in document
- `*` - Match any element
- `text()` - Get text content
- `[predicate]` - Filter elements

Examples:
```json
{
  "id": "123",                    // /id
  "data": {                       // /data/id
    "id": "456"
  },
  "items": [                      // /items/*/id
    {"id": "789"},
    {"id": "012"}
  ],
  "user": {                       // //id (finds any id)
    "profile": {
      "id": "345"
    }
  }
}
```

#### XML Path Syntax

For XML requests, use standard XPath syntax:
- `/` - Start from root
- `//` - Search anywhere in document
- `*` - Match any element
- `text()` - Get text content
- `[predicate]` - Filter elements

Examples:
```xml
<root>
  <id>123</id>                    <!-- /id -->
  <data>                          <!-- /data/id -->
    <id>456</id>
  </data>
  <items>                         <!-- /items/*/id -->
    <item><id>789</id></item>
    <item><id>012</id></item>
  </items>
  <user>                          <!-- //id (finds any id) -->
    <profile>
      <id>345</id>
    </profile>
  </user>
</root>
```

## Extraction Process

1. First, Unimock finds the matching section for the request URL
2. For GET/PUT requests:
   - Extracts ID from the last path segment if it looks like an ID
   - For collection paths (no ID), returns empty ID list
3. For POST requests:
   - Tries to extract ID from the configured header
   - Tries to extract IDs from the request body using configured paths
   - If no IDs found in body/headers, tries to extract from path
   - For collection paths without ID, returns error for JSON requests

## Error Handling

- If no matching section is found, returns "no matching section found for path"
- If path pattern matching fails, returns "failed to match path pattern"
- If no IDs found in JSON request, returns "no IDs found in request"
- If request body is invalid, returns appropriate error message

## Examples

### Configuration Example

```yaml
sections:
  users:
    path_pattern: "/users/*"
    header_id_names: ["X-Resource-ID"]
    body_id_paths:
      - "/id"
      - "/data/id"
      - "//id"
      - "/items/*/id"
  
  orders:
    path_pattern: "/orders/*"
    header_id_names: ["X-Order-ID"]
    body_id_paths:
      - "/order/id"
      - "//id"
```

### Request Examples

1. GET request with ID in path:
   ```
   GET /users/123
   ```
   Extracts: `123`

2. POST request with ID in header:
   ```
   POST /users
   X-Resource-ID: 123
   ```
   Extracts: `123`

3. POST request with ID in JSON body:
   ```
   POST /users
   Content-Type: application/json
   
   {
     "id": "123",
     "name": "test"
   }
   ```
   Extracts: `123`

4. POST request with ID in nested JSON:
   ```
   POST /users
   Content-Type: application/json
   
   {
     "data": {
       "id": "123",
       "name": "test"
     }
   }
   ```
   Extracts: `123`

5. POST request with IDs in array:
   ```
   POST /users
   Content-Type: application/json
   
   {
     "items": [
       {"id": "123"},
       {"id": "456"}
     ]
   }
   ```
   Extracts: `123`, `456` 
