# Unimock Requirements

## Overview
Unimock is a mock service that handles HTTP requests and provides mock responses based on configured scenarios. The system supports various HTTP methods, content types, and ID extraction mechanisms.

## Core Requirements

### 1. Request Handling
- Must handle standard HTTP methods (GET, POST, PUT, DELETE)
- Must support both individual resource and collection endpoints
- Must validate request content types
- Must handle non-existent resources appropriately
- Must support path-based routing

### 2. Resource Management

#### GET Requests
- Must first try to get resource by ID (last part of path)
- If resource not found by ID, must try to get resources collection by the exact path (e.g., /users/999)
- Must NOT fall back to parent collection (e.g., /users) if neither resource nor collection at the requested path exists
- Must return collection responses as a JSON array, including only resources with content type containing 'json'
- Must return 404 Not Found if neither resource nor collection is found
- Must return proper Content-Type headers
- Must sort collection responses by path for consistent ordering
- Must handle malformed collection paths
- Must support case-sensitive path matching
- Must handle trailing slashes in paths
- Must handle mixed content types in collection responses:
  - Must preserve original content for all items (including JSON and binary data)
  - Must return collection responses as application/json
  - Must handle future XML content type support

#### POST Requests
- Must return 201 Created for new resources, including when an ID is auto-generated.
- If no ID is extractable from the request (header, body, or path), a new UUID must be generated and used as the resource ID.
- Must validate request body
- Must support ID extraction from multiple sources
- Must handle duplicate resource creation attempts
- Must support Location header for created resources, pointing to `/path/to/collection/<id>` (where `<id>` can be provided or auto-generated).
- Must handle empty request bodies
- Must handle malformed JSON/XML
- Must handle missing required fields
- Must handle invalid content types

#### PUT Requests
- Must first try to update resource by ID (last part of path)
- Must return 404 Not Found if resource not found
- Must validate request body
- Must maintain resource consistency
- Must handle partial updates
- Must handle empty request bodies
- Must handle malformed JSON/XML
- Must handle missing required fields
- Must handle invalid content types

#### DELETE Requests
- Must first try to delete resource by ID (last part of path)
- If resource not found by ID, must try to delete all resources by the exact path (e.g., /users/999)
- Must NOT fall back to parent collection (e.g., /users) if neither resource nor collection at the requested path exists
- Must return 404 Not Found if neither resource nor collection is found
- Must handle path-based deletion
- Must handle recursive deletion
- Must handle non-existent paths
- Must handle malformed paths
- Must handle case-sensitive path matching

### 3. ID Extraction

#### GET Requests
- Must extract ID from last path segment
- Must handle numeric IDs
- Must handle UUIDs
- Must handle special characters in IDs
- Must handle missing IDs in collection paths
- Must handle malformed path segments

#### POST Requests
- Must extract ID from:
  - JSON/XML body (simple path: `/id`)
  - JSON/XML body (nested path: `/data/id`)
  - JSON/XML body (deep nested path: `//id`)
  - HTTP headers (configurable names)
- Must handle missing IDs: if no ID is found via configured methods, a new UUID is generated and used for resource creation.
- Must handle duplicate IDs
- Must handle invalid ID formats
- Must handle multiple ID paths in configuration
- Must handle case-sensitive ID matching

#### PUT Requests
- Must extract ID from:
  - Last path segment
  - JSON/XML body (if different from path)
  - HTTP headers (if configured)
- Must handle ID mismatch between path and body
- Must handle missing IDs
- Must handle invalid ID formats
- Must handle case-sensitive ID matching

#### DELETE Requests
- Must extract ID from:
  - Last path segment
  - HTTP headers (if configured)
- Must handle missing IDs
- Must handle invalid ID formats
- Must handle case-sensitive ID matching

### 4. Content Type Support
- Must support any content type for requests and responses
- Must return the original Content-Type for single resource responses
- Must return application/json for collection responses
- Must not validate or restrict content types
- Must handle content type with charset and parameters

### 5. Configuration
Must support configuration for:
- Path patterns
- ID extraction paths
- Header names for ID extraction
- Section-specific settings
- Case sensitivity settings
- Default content types
- Error response formats
- Custom status codes
- Response headers

### 6. Error Handling
Must handle:
- Invalid requests
- Resource not found scenarios
- Duplicate resource creation
- Invalid content types
- Invalid JSON/XML bodies
- Missing required fields
- Malformed paths
- Invalid HTTP methods
- Missing headers
- Invalid header values
- Timeout scenarios
- Concurrent modification conflicts
- Storage errors
- Configuration errors

### 7. Storage Requirements
- Must support concurrent access
- Must maintain data consistency
- Must support path-based retrieval
- Must handle multiple resources per path
- Must support efficient deletion of resources
- Must handle storage capacity limits
- Must handle storage errors gracefully
- Must support atomic operations
- Must maintain data integrity
- Must handle storage initialization errors

### 8. Response Formatting
- Must return collection responses as a JSON array
- Must not marshal or format individual items in the array (just concatenate as raw bytes, e.g. [item1, item2])
- Must preserve original content for all items (including JSON, XML, binary, text, etc.)
- Must not validate or restrict content types
- Must return all types of resources, including JSON, XML, binary, and text, as a slice of bytes without marshalling or formatting

### 9. Security Considerations
- Must validate input data
- Must handle malformed requests
- Must prevent resource leakage
- Must validate content types
- Must sanitize response data
- Must handle request size limits
- Must prevent path traversal
- Must handle injection attempts
- Must validate header values
- Must handle authentication headers
- Must handle authorization headers
- Must prevent sensitive data exposure

### 10. Performance Requirements
- Must handle concurrent requests
- Must maintain consistent response times
- Must efficiently handle large collections
- Must optimize storage operations
- Must support caching where appropriate
- Must handle high request rates
- Must handle large request bodies
- Must handle large response bodies
- Must handle connection limits
- Must handle resource limits
- Must handle timeout scenarios
- Must handle rate limiting

## Configuration Examples

### Path Patterns
```yaml
sections:
  users:
    pathPattern: "/users/*"
    bodyIDPaths: ["/id"]
    caseSensitive: true
    defaultContentType: "application/json"
  orders:
    pathPattern: "/orders/*"
    headerIDName: "X-Order-ID"
    bodyIDPaths: ["/orderId", "/order/id"]
    caseSensitive: false
    defaultContentType: "application/json"
```

### ID Extraction Paths
- Simple: `/id`
- Nested: `/data/id`
- Deep nested: `//id`
- Multiple paths per section
- Case-sensitive matching
- Custom header names
- Multiple header support

## Test Coverage
The system includes comprehensive test coverage for:
- Basic CRUD operations
- ID extraction from various sources
- Error handling scenarios
- Content type handling
- Collection operations
- Path-based operations
- Concurrent access
- Edge cases and error conditions
- Performance scenarios
- Security scenarios
- Configuration validation
- Response formatting
- Storage operations
- Header handling
- Path handling
- Content type handling
- Error scenarios
- Timeout scenarios
- Resource limits
- Concurrent access patterns
