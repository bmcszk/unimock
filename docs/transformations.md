# Data Transformations

Transform and modify API responses on-the-fly with Unimock's transformation features.

## What are Transformations?

Transformations let you:
- Modify stored data before returning it
- Add computed fields (timestamps, IDs)
- Transform data formats (JSON ↔ XML)
- Simulate real API behavior (pagination, filtering)

## Basic Transformations

### Adding Timestamps

```yaml
sections:
  - path: "/api/users/*"
    id_path: "/id"
    return_body: true
    transformations:
      - type: "add_timestamp"
        field: "created_at"
        format: "2006-01-02T15:04:05Z"
```

When you POST data, Unimock automatically adds a `created_at` field:

```bash
# POST
curl -X POST http://localhost:8080/api/users \
  -d '{"id": "123", "name": "John"}'

# GET returns
{
  "id": "123",
  "name": "John",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Adding Auto-Generated IDs

```yaml
sections:
  - path: "/api/orders/*"
    id_path: "/order_id"
    transformations:
      - type: "auto_id"
        field: "order_id"
        format: "uuid"  # or "sequence", "timestamp"
```

### Data Enrichment

```yaml
sections:
  - path: "/api/users/*"
    id_path: "/id"
    transformations:
      - type: "enrich"
        add_fields:
          status: "active"
          version: "1.0"
          environment: "test"
```

## Response Modifications

### Filtering Fields

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      - type: "filter_fields"
        include: ["id", "name", "email"]  # Only return these fields
        # OR
        exclude: ["password", "secret"]   # Hide sensitive fields
```

### Field Mapping

```yaml
sections:
  - path: "/api/legacy/*"
    transformations:
      - type: "map_fields"
        mappings:
          user_id: "id"           # Rename user_id to id
          full_name: "name"       # Rename full_name to name
          email_address: "email"  # Rename email_address to email
```

### Default Values

```yaml
sections:
  - path: "/api/products/*"
    transformations:
      - type: "defaults"
        fields:
          currency: "USD"
          in_stock: true
          category: "general"
```

## Advanced Transformations

### Conditional Logic

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      - type: "conditional"
        condition: "/age >= 18"
        then:
          - type: "add_field"
            field: "can_vote"
            value: true
        else:
          - type: "add_field"
            field: "can_vote"
            value: false
```

### Computed Fields

```yaml
sections:
  - path: "/api/orders/*"
    transformations:
      - type: "compute"
        field: "total_with_tax"
        expression: "/total * 1.1"  # Add 10% tax
      - type: "compute"
        field: "full_name"
        expression: "/first_name + ' ' + /last_name"
```

### Array Transformations

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      - type: "array_transform"
        field: "orders"
        operations:
          - type: "sort"
            by: "created_at"
            order: "desc"
          - type: "limit"
            count: 10
```

## Format Transformations

### JSON to XML

```yaml
sections:
  - path: "/api/xml/*"
    transformations:
      - type: "format_convert"
        from: "json"
        to: "xml"
        root_element: "response"
```

Example:
```bash
# POST JSON
curl -X POST http://localhost:8080/api/xml/users \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "John"}'

# GET returns XML
curl http://localhost:8080/api/xml/users/123
# Returns:
# <response>
#   <id>123</id>
#   <name>John</name>
# </response>
```

### Response Wrapping

```yaml
sections:
  - path: "/api/v2/*"
    transformations:
      - type: "wrap_response"
        wrapper:
          success: true
          data: "${response}"
          timestamp: "${now}"
```

Returns:
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "John"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Dynamic Transformations

### Based on Request Headers

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      - type: "header_based"
        rules:
          - header: "Accept-Language"
            value: "es"
            transformations:
              - type: "translate"
                fields: ["name", "description"]
                to_language: "spanish"
          - header: "X-API-Version"
            value: "v2"
            transformations:
              - type: "add_field"
                field: "api_version"
                value: "2.0"
```

### Based on Query Parameters

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      - type: "query_based"
        rules:
          - param: "include_details"
            value: "true"
            transformations:
              - type: "enrich"
                add_fields:
                  last_login: "2024-01-15T10:30:00Z"
                  permissions: ["read", "write"]
```

## Pagination Simulation

```yaml
sections:
  - path: "/api/users"  # List endpoint
    transformations:
      - type: "paginate"
        page_size: 10
        page_param: "page"
        size_param: "size"
        response_format:
          data: "${items}"
          page: "${current_page}"
          total_pages: "${total_pages}"
          total_items: "${total_count}"
```

Usage:
```bash
# GET /api/users?page=2&size=5
{
  "data": [...],
  "page": 2,
  "total_pages": 20,
  "total_items": 100
}
```

## Error Simulation

### Random Errors

```yaml
sections:
  - path: "/api/flaky/*"
    transformations:
      - type: "error_simulation"
        probability: 0.1  # 10% chance of error
        errors:
          - status: 500
            message: "Internal server error"
          - status: 503
            message: "Service unavailable"
```

### Conditional Errors

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      - type: "conditional_error"
        condition: "/id == 'invalid'"
        error:
          status: 400
          message: "Invalid user ID"
```

## Performance Simulation

### Response Delays

```yaml
sections:
  - path: "/api/slow/*"
    transformations:
      - type: "delay"
        min_ms: 100
        max_ms: 500  # Random delay between 100-500ms
```

### Size Limits

```yaml
sections:
  - path: "/api/upload/*"
    transformations:
      - type: "size_check"
        max_bytes: 1048576  # 1MB limit
        error_status: 413
        error_message: "Request too large"
```

## Transformation Chaining

Combine multiple transformations:

```yaml
sections:
  - path: "/api/users/*"
    transformations:
      # 1. Add timestamp
      - type: "add_timestamp"
        field: "created_at"
      
      # 2. Add computed field
      - type: "compute"
        field: "display_name"
        expression: "/first_name + ' ' + /last_name"
      
      # 3. Filter sensitive data
      - type: "filter_fields"
        exclude: ["password", "ssn"]
      
      # 4. Wrap response
      - type: "wrap_response"
        wrapper:
          success: true
          data: "${response}"
```

## Testing Transformations

### Validate Transformation Output

```bash
#!/bin/bash

# POST test data
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "test", "first_name": "John", "last_name": "Doe"}'

# GET and verify transformations
response=$(curl -s http://localhost:8080/api/users/test)

# Check if timestamp was added
echo "$response" | jq '.created_at' || echo "❌ Missing created_at"

# Check if computed field exists
echo "$response" | jq '.display_name' || echo "❌ Missing display_name"

# Check if sensitive fields were filtered
echo "$response" | jq '.password' && echo "❌ Password not filtered"

echo "✅ Transformations working correctly"
```

## Best Practices

### 1. Keep Transformations Simple

```yaml
# Good - simple, clear transformation
transformations:
  - type: "add_timestamp"
    field: "created_at"

# Avoid - complex, hard to debug
transformations:
  - type: "conditional"
    condition: "complex expression here"
    then: [multiple nested transformations]
```

### 2. Test Each Transformation

Test transformations individually before chaining them.

### 3. Document Complex Logic

```yaml
sections:
  - path: "/api/orders/*"
    # This transformation calculates tax based on customer location
    transformations:
      - type: "compute"
        field: "tax_amount"
        expression: "/total * (/customer/state == 'CA' ? 0.1 : 0.05)"
```

### 4. Use Consistent Field Names

```yaml
# Use consistent timestamp formats across all endpoints
transformations:
  - type: "add_timestamp"
    field: "created_at"
    format: "2006-01-02T15:04:05Z"  # ISO 8601
```

### 5. Handle Edge Cases

```yaml
transformations:
  - type: "compute"
    field: "display_name"
    expression: "/first_name + ' ' + /last_name"
    default: "Unknown User"  # Fallback for missing names
```

Transformations make Unimock responses more realistic and help simulate real API behavior in your tests.