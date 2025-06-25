# Technical Endpoints

Unimock provides a set of technical endpoints for monitoring and operations under the `/_uni/` path prefix.

## Health Check

The health check endpoint returns the current status and uptime of the server.

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

## Metrics

The metrics endpoint provides statistics about the server usage.

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

## Scenarios

Unimock provides a RESTful API for managing test scenarios. Scenarios can be created, retrieved, updated, and deleted via the `/_uni/scenarios` endpoint.

### Scenario Fields

Each scenario contains the following fields:

- `uuid`: Unique identifier for the scenario (auto-generated if not provided)
- `method`: HTTP method (GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS)
- `path`: API path this scenario is for (supports wildcards with `*`)
- `status_code`: HTTP status code to return
- `content_type`: Content type of the response
- `location`: Location header to return (optional)
- `headers`: Additional response headers (optional)
- `data`: Response body data (optional)

### Create a Scenario

```bash
curl -X POST http://localhost:8080/_uni/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/users/123",
    "status_code": 200,
    "content_type": "application/json",
    "data": "{\"id\":\"123\",\"name\":\"John Doe\"}"
  }'
```

Response:
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/users/123",
  "status_code": 200,
  "content_type": "application/json",
  "data": "{\"id\":\"123\",\"name\":\"John Doe\"}"
}
```

### Get a Scenario

```bash
curl -X GET http://localhost:8080/_uni/scenarios/550e8400-e29b-41d4-a716-446655440000
```

### List All Scenarios

```bash
curl -X GET http://localhost:8080/_uni/scenarios
```

### Update a Scenario

```bash
curl -X PUT http://localhost:8080/_uni/scenarios/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/users/123",
    "status_code": 201,
    "content_type": "application/json",
    "data": "{\"id\":\"123\",\"name\":\"Jane Doe\"}"
  }'
```

### Delete a Scenario

```bash
curl -X DELETE http://localhost:8080/_uni/scenarios/550e8400-e29b-41d4-a716-446655440000
``` 
