# Usage Examples

This document provides practical examples of how to use Unimock for different scenarios and request types.

## Basic CRUD Operations

### Create Resource with ID in Body

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "John Doe", "email": "john@example.com"}'
```

### Create Resource with ID in Header

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

### Retrieve Individual Resource

```bash
curl http://localhost:8080/api/users/123
```

### Retrieve Collection

```bash
curl http://localhost:8080/api/users
```

### Update Resource

```bash
curl -X PUT http://localhost:8080/api/users/123 \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "Jane Doe", "email": "jane@example.com"}'
```

### Delete Resource

```bash
curl -X DELETE http://localhost:8080/api/users/123
```

## JSON Examples

### Simple JSON Resource

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user123",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "admin"
  }'
```

### Nested JSON Structure

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user456",
    "profile": {
      "name": "Jane Smith",
      "email": "jane@example.com",
      "preferences": {
        "theme": "dark",
        "notifications": true
      }
    },
    "roles": ["user", "editor"]
  }'
```

### JSON with Arrays

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "id": "order789",
    "customer": "John Doe",
    "items": [
      {"id": "item1", "product": "Laptop", "price": 1200, "quantity": 1},
      {"id": "item2", "product": "Mouse", "price": 25, "quantity": 2}
    ],
    "total": 1250
  }'
```

## XML Examples

### Simple XML Resource

```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/xml" \
  -d '<?xml version="1.0" encoding="UTF-8"?>
<product>
  <id>prod123</id>
  <name>Example Product</name>
  <price>29.99</price>
  <category>Electronics</category>
</product>'
```

### Nested XML Structure

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/xml" \
  -d '<?xml version="1.0" encoding="UTF-8"?>
<order>
  <id>order456</id>
  <customer>
    <name>John Doe</name>
    <email>john@example.com</email>
    <address>
      <street>123 Main St</street>
      <city>Anytown</city>
      <zipcode>12345</zipcode>
    </address>
  </customer>
  <items>
    <item>
      <id>item1</id>
      <product>Widget</product>
      <quantity>2</quantity>
    </item>
  </items>
</order>'
```

## Multiple ID Extraction

### IDs from Different Body Paths

Given this configuration:
```yaml
sections:
  products:
    path_pattern: "/products/*"
    body_id_paths:
      - "/product/sku"
      - "/meta/uuid"
```

Create a resource with multiple IDs:
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "product": {
      "sku": "ABC123",
      "name": "Multi-ID Product"
    },
    "meta": {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "created": "2024-01-15T10:30:00Z"
    }
  }'
```

Retrieve by either ID:
```bash
# Retrieve by SKU
curl http://localhost:8080/products/ABC123

# Retrieve by UUID
curl http://localhost:8080/products/550e8400-e29b-41d4-a716-446655440000
```

### Header and Body ID Combination

Given this configuration:
```yaml
sections:
  items:
    path_pattern: "/api/items/*"
    header_id_names: ["X-Item-Token"]
    body_id_paths:
      - "/itemID"
```

Create with both header and body IDs:
```bash
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -H "X-Item-Token: token789" \
  -d '{
    "itemID": "item456",
    "name": "Dual-ID Item",
    "description": "Item with both header and body IDs"
  }'
```

Retrieve by either ID:
```bash
# Retrieve by token
curl http://localhost:8080/api/items/token789

# Retrieve by item ID
curl http://localhost:8080/api/items/item456
```

## Deep Path Examples

### Nested Resource Creation

```bash
# Create user first
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "user123", "name": "John Doe"}'

# Create order for that user
curl -X POST http://localhost:8080/api/users/user123/orders \
  -H "Content-Type: application/json" \
  -d '{
    "id": "order456",
    "product": "Laptop",
    "price": 1200
  }'
```

### Retrieve Nested Resource

```bash
# Get specific order for specific user
curl http://localhost:8080/api/users/user123/orders/order456

# Get all orders for a user
curl http://localhost:8080/api/users/user123/orders
```

## Different Content Types

### Plain Text

```bash
curl -X POST http://localhost:8080/api/logs/log123 \
  -H "Content-Type: text/plain" \
  -d "Application started successfully at $(date)"
```

### Binary Data

```bash
curl -X POST http://localhost:8080/api/files/image123 \
  -H "Content-Type: image/jpeg" \
  --data-binary @photo.jpg
```

### Custom Content Type

```bash
curl -X POST http://localhost:8080/api/configs/app123 \
  -H "Content-Type: application/yaml" \
  -d 'server:
  port: 8080
  host: localhost
database:
  driver: postgres
  host: db.example.com'
```

## HEAD Method Examples

### Check Resource Existence

```bash
# Check if user exists (returns only headers)
curl -I http://localhost:8080/api/users/123

# Check with verbose output to see all headers
curl -v -I http://localhost:8080/api/users/123
```

## Advanced Examples

### Using Custom Headers

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "X-User-ID: custom123" \
  -H "X-Source: mobile-app" \
  -H "Authorization: Bearer token123" \
  -d '{
    "name": "Mobile User",
    "platform": "iOS"
  }'
```

### Query Parameters (Passed Through)

```bash
# Query parameters are preserved and passed to stored data
curl -X POST "http://localhost:8080/api/users?source=registration&campaign=summer2024" \
  -H "Content-Type: application/json" \
  -d '{"id": "user789", "name": "Campaign User"}'

# Retrieve with same parameters
curl "http://localhost:8080/api/users/user789?source=registration&campaign=summer2024"
```

### Large JSON Payload

```bash
curl -X POST http://localhost:8080/api/documents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc123",
    "title": "Large Document",
    "metadata": {
      "author": "John Doe",
      "created": "2024-01-15T10:30:00Z",
      "tags": ["important", "draft", "review"],
      "permissions": {
        "read": ["user1", "user2", "user3"],
        "write": ["user1"],
        "admin": ["admin1"]
      }
    },
    "content": {
      "sections": [
        {
          "title": "Introduction",
          "body": "This is the introduction section with lots of text..."
        },
        {
          "title": "Main Content",
          "body": "This is the main content section with even more text..."
        }
      ]
    },
    "attachments": [
      {"name": "chart.png", "type": "image", "size": 1024000},
      {"name": "data.csv", "type": "spreadsheet", "size": 2048000}
    ]
  }'
```

## Error Scenarios

### Missing Required ID

```bash
# This will fail if ID extraction is required but no ID found
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "No ID User"}'
# Expected response: 400 Bad Request
```

### Duplicate ID Creation

```bash
# Create first resource
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "duplicate123", "name": "First User"}'

# Try to create another with same ID
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "duplicate123", "name": "Second User"}'
# Expected response: 409 Conflict
```

### Non-existent Resource

```bash
# Try to retrieve non-existent resource
curl http://localhost:8080/api/users/nonexistent
# Expected response: 404 Not Found
```

## Working with Scenarios

### Create and Use Scenario

```bash
# Create a scenario for error testing
curl -X POST http://localhost:8080/_uni/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "uuid": "user-server-error",
    "method": "POST",
    "path": "/api/users",
    "status_code": 500,
    "content_type": "application/json",
    "data": "{\"error\": \"Internal server error\", \"code\": \"SERVER_ERROR\"}"
  }'

# Now POST requests to /api/users will return the scenario response
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "user123", "name": "Test User"}'
# Returns: 500 Internal Server Error with scenario data

# Remove the scenario to restore normal behavior
curl -X DELETE http://localhost:8080/_uni/scenarios/user-server-error
```

These examples demonstrate the flexibility of Unimock in handling various HTTP request types, content formats, and ID extraction patterns.