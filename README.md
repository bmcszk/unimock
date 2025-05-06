# Unimock

A Universal Mock HTTP server designed for e2e testing scenarios where you need to mock third-party REST services. This server can work as a RESTful service, XML service, or any other HTTP-based service.

## Purpose

Unimock was created to solve a common problem in e2e testing: the need to mock third-party REST services that you don't control. Instead of building custom mock services for each test scenario, Unimock provides a universal solution that can:

- Handle any HTTP request type
- Support both JSON and XML formats
- Extract IDs from various sources (body, headers, paths)
- Store data in a thread-safe in-memory map
- Preserve content types and metadata

## Key Features

- **Universal Support**: Works with any HTTP-based service (REST, XML, etc.)
- **Flexible ID Extraction**:
  - XPath-based ID extraction for JSON and XML bodies
  - Header-based ID extraction
  - Path-based ID fallback
- **Thread-Safe Storage**:
  - In-memory map with mutex protection
  - Stores IDs, metadata, content type, and body
  - Path-based storage for data without IDs
- **Full HTTP Support**:
  - All HTTP methods (GET, POST, PUT, DELETE)
  - Custom content types
  - Proper status codes and headers
- **Technical Endpoints**:
  - Health check endpoint (`/_uni/health`)
  - Metrics endpoint (`/_uni/metrics`)
  - Secure path prefix for monitoring and operations

## Requirements

### Path Structure
- Collection paths (e.g., `/users`) - for multiple resources
- Resource paths (e.g., `/users/123`) - for single resources
- Deep paths (e.g., `/users/123/orders/456`) - treated as single resources
- Last path segment is used as ID when no ID found in body/headers
- All paths are stored without trailing slashes

### Storage
- In-memory storage with thread-safe operations
- Support for storing data with multiple external IDs
- Support for path-based storage and retrieval
- UUID-based internal storage IDs
- Persistence of external ID to storage ID mapping until explicit deletion
- Paths are stored without trailing slashes
- Location field stores the full resource path with ID

### HTTP Handler

#### GET Requests
- Single item retrieval by ID:
  - ID is the last segment of the path (e.g., `/users/123` -> ID is "123")
  - Returns the item with its original content type
  - Location header contains the full resource path
- Array retrieval by path:
  - Returns all items stored at the given path as a JSON array
  - Always returns JSON array format, regardless of item content types
- No request body parsing for GET requests
- Returns 404 if neither ID nor path match

#### POST Requests
- Creates new resources
- Accepts any path structure (collection, resource, or deep path)
- Extracts IDs from:
  1. Headers (if configured)
  2. Request body:
     - For JSON: Must have ID in body (no path fallback)
     - For XML: Uses body ID or falls back to last path segment
     - For others: Uses last path segment as ID
- Returns 409 if resource already exists
- Returns 201 on successful creation with:
  - Location header containing the full resource path from data.Location
  - Created resource in response body

#### PUT Requests
- Updates existing resources
- Uses GET-style ID extraction:
  - ID from last path segment
  - No body parsing
  - Consistent with GET behavior
- Returns 404 for non-existent resources
- Location header contains the full resource path from data.Location
- Example paths:
  - `/users/123` -> updates resource with ID "123"
  - `/users/123/orders/456` -> updates resource with ID "456"

#### DELETE Requests
- Two-step deletion process:
  1. Try ID-based deletion (GET logic)
     - Example: `/users/123` -> deletes resource with ID "123"
  2. Fall back to path-based deletion
     - Example: If ID not found, deletes all resources under `/users/123/*`
- Returns 204 on success
- Returns 404 if no resources found
- Location header contains the full resource path from data.Location
- Example paths:
  - `/users/123` -> first tries to delete resource with ID "123", then falls back to deleting all resources under `/users/123/*`
  - `/users/123/orders` -> first tries to delete resource with ID "orders", then falls back to deleting all resources under `/users/123/orders/*`

### ID Extraction
- Configurable ID paths for JSON and XML bodies
- Configurable ID header name
- Path-based fallback for non-JSON requests
- No body parsing for GET requests
- Support for various ID types:
  - Numbers (e.g., "123", "42")
  - Text (e.g., "user-123", "order-xyz")
  - UUIDs (e.g., "550e8400-e29b-41d4-a716-446655440000")
  - Dates (e.g., "2024-03-20", "2024-03-20T15:30:00Z")

### Content Types
- Supports any content type for storage
- JSON array responses for path-based GET requests
- Original content type preserved for ID-based GET requests

## Assumptions
1. GET requests never have a body
2. Path-based GET requests always return JSON arrays
3. External ID to storage ID mapping persists until explicit deletion
4. Multiple external IDs can map to the same storage ID
5. Storage operations are thread-safe
6. Content types are preserved for individual items
7. JSON array responses are used for path-based queries regardless of item content types
8. JSON requests must have ID in body or header
9. Non-JSON requests can use last path segment as ID

## Features

- Supports all HTTP methods (GET, POST, PUT, DELETE, etc.)
- In-memory data storage with thread-safe operations
- Automatic ID extraction from:
  - Request headers
  - JSON body using XPath expressions
  - XML body using XPath expressions
  - Last path segment (for non-JSON requests)
- Configurable ID extraction paths
- Stores metadata, content type, and body for each request
- Supports deep paths with proper ID extraction

## Installation

```bash
go get github.com/bmcszk/unimock
```

## Usage

To run the mock server:

```bash
make run
```

This will start the server on port 8080.

## Technical Endpoints

Unimock provides a set of technical endpoints for monitoring and operations under the `/_uni/` path prefix:

### Health Check

```bash
curl -X GET http://localhost:8080/_uni/health
```

Response:
```json
{
  "status": "ok",
  "uptime": "1h23m45s"
}
```

### Metrics

```bash
curl -X GET http://localhost:8080/_uni/metrics
```

Response:
```json
{
  "request_count": 42,
  "api_endpoints": {
    "/_uni/health": 3,
    "/_uni/metrics": 2,
    "/api/users": 20,
    "/api/users/123": 17
  }
}
```

## Example Requests

### Store data with ID in header
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-Resource-ID: 123" \
  -d '{"name": "test"}' \
  http://localhost:8080/users
```

### Store data with ID in body
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "test"}' \
  http://localhost:8080/users
```

### Store data with deep path
```bash
curl -X POST \
  -H "Content-Type: application/xml" \
  -d '<order><name>test</name></order>' \
  http://localhost:8080/users/123/orders/456
```

### Retrieve data
```bash
# Get single resource
curl -X GET http://localhost:8080/users/123

# Get collection
curl -X GET http://localhost:8080/users

# Get deep resource
curl -X GET http://localhost:8080/users/123/orders/456
```

### Delete data
```bash
curl -X DELETE http://localhost:8080/users/123
```

## Configuration

The server can be configured with:

- `idPaths`: List of XPath expressions to find IDs in request bodies
- `idHeader`: Header name to look for ID in request headers

### YAML Configuration
- Config file using yaml
- Config separated into sections
- Every section has path pattern
- Every section has paths to id in body
- Every section has header name for id in header

## License

This project is licensed under the [MIT License](LICENSE) - see the [LICENSE](LICENSE) file for details.

## Technical Details

### ID Extraction
- Uses [jsonquery](https://github.com/antchfx/jsonquery) for JSON parsing
- Uses [xmlquery](https://github.com/antchfx/xmlquery) for XML parsing
- Configurable XPath expressions for ID location
- Multiple ID support in single request
- Header-based ID extraction option
- Path-based fallback for non-JSON requests

### Storage
- Thread-safe in-memory map
- Stores:
  - External IDs
  - Content type
  - Request body
  - Path information (without trailing slashes)
  - Location field with full resource path
- Path-based storage for requests without IDs

### HTTP Methods

#### GET
- Single resource retrieval by ID
- Collection retrieval by path
- Returns original content type for single resources
- Returns JSON array for collections

#### POST
- Creates new resources
- Extracts IDs from:
  1. Headers (if configured)
  2. Request body (JSON/XML)
  3. Path (fallback)
- Returns 201 with Location header

#### PUT
- Updates existing resources
- Uses GET-style ID extraction:
  - ID from last path segment
  - No body parsing
  - Consistent with GET behavior
- Returns 404 for non-existent resources
- Location header contains the full resource path from data.Location
- Example paths:
  - `/users/123` -> updates resource with ID "123"
  - `/users/123/orders/456` -> updates resource with ID "456"

#### DELETE
- Two-step deletion process:
  1. Try ID-based deletion (GET logic)
     - Example: `/users/123` -> deletes resource with ID "123"
  2. Fall back to path-based deletion
     - Example: If ID not found, deletes all resources under `/users/123/*`
- Returns 204 on success
- Returns 404 if no resources found
- Location header contains the full resource path from data.Location
- Example paths:
  - `/users/123` -> first tries to delete resource with ID "123", then falls back to deleting all resources under `/users/123/*`
  - `/users/123/orders` -> first tries to delete resource with ID "orders", then falls back to deleting all resources under `/users/123/orders/*`

## Use Cases

1. **E2E Testing**: Mock third-party services in your test environment
2. **Development**: Simulate external services during development
3. **Integration Testing**: Test your application's integration with external services
4. **API Development**: Prototype and test API designs

## Best Practices

1. Configure appropriate ID extraction paths for your use case
2. Use meaningful paths that reflect your API structure
3. Include proper content types in requests
4. Handle both success and error cases in your tests
5. Clean up test data after use

### Path Matching and ID Extraction
- Every request coming to server should match one of the configured path patterns
- Every request may have different ID paths and header name to extract ID taken from config
- Path patterns are matched in order of configuration
- ID extraction follows the order: header name -> body paths -> path segment

## Documentation

- [ID Extraction](docs/id_extraction.md) - Learn how to configure ID extraction from requests
