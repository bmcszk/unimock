# Scenarios

Scenarios allow you to create predefined responses that override normal mock behavior. They are useful for testing specific edge cases, error conditions, or complex response patterns.

## What are Scenarios?

Scenarios are predefined mock responses that take precedence over normal resource storage. When a request matches a scenario's path and method, Unimock returns the scenario's predefined response instead of looking up stored data.

**Key features:**
- **Method + Path matching** - Scenarios match specific HTTP method and path combinations
- **Override behavior** - Scenarios bypass normal mock storage lookup
- **Flexible responses** - Custom status codes, headers, and response data
- **CRUD management** - Create, read, update, and delete scenarios via API
- **File-based loading** - Load scenarios from YAML files at startup

## Creating Scenarios

### Via API (Runtime)

Create scenarios dynamically using the REST API:

```bash
curl -X POST http://localhost:8080/_uni/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "uuid": "user-not-found",
    "method": "GET",
    "path": "/api/users/999",
    "status_code": 404,
    "content_type": "application/json",
    "data": "{\"error\": \"User not found\", \"code\": \"USER_NOT_FOUND\"}"
  }'
```

Response:
```json
{
  "uuid": "user-not-found",
  "method": "GET", 
  "path": "/api/users/999",
  "status_code": 404,
  "content_type": "application/json",
  "data": "{\"error\": \"User not found\", \"code\": \"USER_NOT_FOUND\"}"
}
```

### Via YAML Configuration (Startup)

Load scenarios from a YAML file when starting Unimock:

**scenarios.yaml:**
```yaml
scenarios:
  # GET scenario for a specific user
  - uuid: "example-user-001"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "role": "admin"
      }

  # Error scenario for non-existent resource
  - uuid: "example-user-not-found"
    method: "GET"
    path: "/api/users/999"
    status_code: 404
    content_type: "application/json"
    data: |
      {
        "error": "User not found",
        "code": "USER_NOT_FOUND",
        "details": "User with ID 999 does not exist"
      }

  # POST scenario with custom location
  - uuid: "example-user-create"
    method: "POST"
    path: "/api/users"
    status_code: 201
    content_type: "application/json"
    location: "/api/users/456"
    headers:
      X-User-ID: "456"
    data: |
      {
        "id": "456",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "role": "user"
      }
```

**Load scenarios at startup:**
```bash
# Using environment variable
UNIMOCK_SCENARIOS_FILE=scenarios.yaml unimock

# Using Docker
docker run -p 8080:8080 \
  -v $(pwd)/scenarios.yaml:/etc/unimock/scenarios.yaml \
  -e UNIMOCK_SCENARIOS_FILE=/etc/unimock/scenarios.yaml \
  ghcr.io/bmcszk/unimock:latest
```

## Scenario Fields

| Field | Required | Description |
|-------|----------|-------------|
| `uuid` | No | Unique identifier (auto-generated if not provided) |
| `method` | Yes | HTTP method (GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS) |
| `path` | Yes | Request path to match (supports wildcards with `*`) |
| `status_code` | Yes | HTTP status code to return |
| `content_type` | Yes | Response content type |
| `data` | No | Response body data |
| `location` | No | Location header value |
| `headers` | No | Additional response headers |

### Path Matching

Scenarios support wildcard path matching:

```yaml
scenarios:
  # Exact path match
  - method: "GET"
    path: "/api/users/123"
    # matches only: GET /api/users/123

  # Wildcard match
  - method: "GET"
    path: "/api/orders/*"
    # matches: GET /api/orders/456, GET /api/orders/789, etc.
```

## Managing Scenarios

### List All Scenarios

```bash
curl http://localhost:8080/_uni/scenarios
```

### Get Specific Scenario

```bash
curl http://localhost:8080/_uni/scenarios/user-not-found
```

### Update Scenario

```bash
curl -X PUT http://localhost:8080/_uni/scenarios/user-not-found \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/users/999",
    "status_code": 410,
    "content_type": "application/json",
    "data": "{\"error\": \"User permanently deleted\"}"
  }'
```

### Delete Scenario

```bash
curl -X DELETE http://localhost:8080/_uni/scenarios/user-not-found
```

## Common Use Cases

### Error Testing

Create scenarios for various error conditions:

```yaml
scenarios:
  # 404 Not Found
  - uuid: "user-not-found"
    method: "GET"
    path: "/api/users/999"
    status_code: 404
    content_type: "application/json"
    data: '{"error": "User not found"}'

  # 500 Server Error
  - uuid: "server-error"
    method: "POST"
    path: "/api/users"
    status_code: 500
    content_type: "application/json"
    data: '{"error": "Internal server error"}'

  # 400 Bad Request
  - uuid: "validation-error"
    method: "POST"
    path: "/api/users"
    status_code: 400
    content_type: "application/json"
    data: '{"error": "Invalid email format"}'
```

### Authentication/Authorization Testing

```yaml
scenarios:
  # 401 Unauthorized
  - uuid: "unauthorized"
    method: "GET"
    path: "/api/secure/*"
    status_code: 401
    content_type: "application/json"
    data: '{"error": "Authentication required"}'

  # 403 Forbidden
  - uuid: "forbidden"
    method: "DELETE"
    path: "/api/admin/*"
    status_code: 403
    content_type: "application/json"
    data: '{"error": "Insufficient permissions"}'
```

### Different Content Types

```yaml
scenarios:
  # XML Response
  - uuid: "product-xml"
    method: "GET"
    path: "/products/abc123"
    status_code: 200
    content_type: "application/xml"
    data: |
      <?xml version="1.0" encoding="UTF-8"?>
      <product>
        <sku>abc123</sku>
        <name>Example Product</name>
        <price>29.99</price>
      </product>

  # Plain Text Response
  - uuid: "health-text"
    method: "GET"
    path: "/health"
    status_code: 200
    content_type: "text/plain"
    data: "OK"

  # Empty Response with Headers
  - uuid: "created-empty"
    method: "POST"
    path: "/api/tasks"
    status_code: 201
    content_type: "application/json"
    location: "/api/tasks/456"
    headers:
      X-Task-ID: "456"
    data: ""
```

### HEAD Method Support

```yaml
scenarios:
  # HEAD scenario for checking resource existence
  - uuid: "user-head-check"
    method: "HEAD"
    path: "/api/users/789"
    status_code: 200
    content_type: "application/json"
    headers:
      X-User-ID: "789"
      Last-Modified: "Wed, 21 Oct 2015 07:28:00 GMT"
```

## Scenario Priority

Scenarios take precedence over normal mock storage:

1. **Scenario match** - If a request matches a scenario's method and path, return the scenario response
2. **Normal storage** - If no scenario matches, use normal mock storage lookup
3. **404 Not Found** - If neither scenario nor stored data exists, return 404

Example behavior:
```bash
# 1. Create a scenario
curl -X POST http://localhost:8080/_uni/scenarios \
  -d '{"method": "GET", "path": "/api/users/123", "status_code": 418, "data": "I am a teapot"}'

# 2. Store normal data
curl -X POST http://localhost:8080/api/users \
  -d '{"id": "123", "name": "John Doe"}'

# 3. GET request returns scenario (not stored data)
curl http://localhost:8080/api/users/123
# Returns: 418 I'm a teapot, "I am a teapot"

# 4. Delete scenario
curl -X DELETE http://localhost:8080/_uni/scenarios/{scenario-id}

# 5. GET request now returns stored data
curl http://localhost:8080/api/users/123  
# Returns: 200 OK, {"id": "123", "name": "John Doe"}
```

## Integration with Testing

### Go Testing

```go
func TestUserNotFound(t *testing.T) {
    client := client.NewClient("http://localhost:8080")
    
    // Create scenario for user not found
    scenario := &model.Scenario{
        Method:      "GET",
        Path:        "/api/users/999",
        StatusCode:  404,
        ContentType: "application/json",
        Data:        `{"error": "User not found"}`,
    }
    
    err := client.CreateScenario(context.Background(), scenario)
    require.NoError(t, err)
    
    // Test the scenario
    resp, err := client.Get(context.Background(), "/api/users/999", nil)
    require.NoError(t, err)
    assert.Equal(t, 404, resp.StatusCode)
    
    // Cleanup
    err = client.DeleteScenario(context.Background(), scenario.UUID)
    require.NoError(t, err)
}
```

### JavaScript Testing

```javascript
test('should handle user not found', async () => {
    // Create scenario
    await fetch('http://localhost:8080/_uni/scenarios', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            method: 'GET',
            path: '/api/users/999',
            status_code: 404,
            content_type: 'application/json',
            data: '{"error": "User not found"}'
        })
    });
    
    // Test the scenario
    const response = await fetch('http://localhost:8080/api/users/999');
    expect(response.status).toBe(404);
    
    const data = await response.json();
    expect(data.error).toBe('User not found');
});
```

## Best Practices

### 1. Use Descriptive UUIDs

```yaml
# Good - descriptive UUID
- uuid: "user-not-found-404"
  method: "GET"
  path: "/api/users/999"

# Avoid - generic UUID  
- uuid: "scenario1"
  method: "GET"
  path: "/api/users/999"
```

### 2. Clean Up Test Scenarios

Always delete scenarios created during tests to avoid interference:

```go
defer func() {
    client.DeleteScenario(context.Background(), scenarioID)
}()
```

### 3. Use Wildcards for Pattern Matching

```yaml
# Match all order endpoints
- uuid: "orders-processing"
  method: "GET"
  path: "/api/orders/*"
  status_code: 200
  data: '{"status": "processing"}'
```

### 4. Document Scenario Purpose

```yaml
scenarios:
  # Test scenario: Simulate payment gateway timeout
  - uuid: "payment-timeout"
    method: "POST"
    path: "/api/payments"
    status_code: 504
    content_type: "application/json"
    data: '{"error": "Gateway timeout"}'
```

Scenarios provide powerful control over your mock responses, enabling comprehensive testing of both happy paths and edge cases.