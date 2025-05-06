# Usage Examples

This document provides examples of how to use Unimock for different scenarios.

## Basic Examples

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

### Update data

```bash
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "updated"}' \
  http://localhost:8080/users/123
```

### Delete data

```bash
curl -X DELETE http://localhost:8080/users/123
```

## Advanced Examples

### Working with JSON

#### Create nested JSON data

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "id": "123",
    "user": {
      "name": "John Doe",
      "email": "john@example.com",
      "profile": {
        "age": 30,
        "occupation": "Developer"
      }
    }
  }' \
  http://localhost:8080/users
```

#### Create data with array

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "id": "123",
    "name": "John Doe",
    "orders": [
      {"id": "order1", "product": "Laptop", "price": 1200},
      {"id": "order2", "product": "Phone", "price": 800}
    ]
  }' \
  http://localhost:8080/users
```

### Working with XML

#### Create XML data

```bash
curl -X POST \
  -H "Content-Type: application/xml" \
  -d '<user>
    <id>123</id>
    <name>John Doe</name>
    <email>john@example.com</email>
  </user>' \
  http://localhost:8080/users
```

#### Create nested XML data

```bash
curl -X POST \
  -H "Content-Type: application/xml" \
  -d '<user>
    <id>123</id>
    <profile>
      <name>John Doe</name>
      <email>john@example.com</email>
      <address>
        <street>123 Main St</street>
        <city>Anytown</city>
      </address>
    </profile>
  </user>' \
  http://localhost:8080/users
```

### Working with Scenarios

#### Create a scenario

```bash
curl -X POST http://localhost:8080/_uni/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/users/123",
    "statusCode": 200,
    "contentType": "application/json",
    "data": "{\"id\":\"123\",\"name\":\"John Doe\"}"
  }'
```

#### Update a scenario

```bash
curl -X PUT http://localhost:8080/_uni/scenarios/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/users/123",
    "statusCode": 201,
    "contentType": "application/json",
    "data": "{\"id\":\"123\",\"name\":\"Jane Doe\"}"
  }'
```

### Custom Content Types

#### Create binary data

```bash
curl -X POST \
  -H "Content-Type: application/octet-stream" \
  --data-binary @file.bin \
  http://localhost:8080/files/file1
```

#### Create text data

```bash
curl -X POST \
  -H "Content-Type: text/plain" \
  -d 'This is a plain text file.' \
  http://localhost:8080/files/file1
``` 
