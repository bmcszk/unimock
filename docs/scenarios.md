# Scenario System

## Overview

The Scenario System provides a way to create and manage test scenarios within Unimock using a dedicated in-memory database.

## Key Features

- **Simple Storage**: Dedicated ID-to-scenario map
- **REST API**: Full CRUD operations via HTTP endpoints
- **Thread Safety**: Concurrent access protection

## Scenario-Based Mocking

The scenario system allows for predefined mock responses based on request paths. When a request is made to a path that matches a scenario's defined path, the mock server will return the predefined response instead of using the regular mock functionality.

### Creating a Scenario for Path-Based Mocking

```bash
# Create a scenario for /api/users path
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/users",
    "statusCode": 200,
    "contentType": "application/json",
    "data": "[{\"id\": 1, \"name\": \"John\"}, {\"id\": 2, \"name\": \"Jane\"}]"
  }' \
  http://localhost:8080/_uni/scenarios
```

### Using Path-Based Scenarios

When a request is made to the path defined in a scenario, the mock server will return the predefined response:

```bash
# This will return the predefined scenario data
curl -X GET http://localhost:8080/api/users
```

This feature is particularly useful for:
- Setting up consistent test environments
- Mocking specific API endpoints with predefined responses
- Creating complex test scenarios with minimal setup

## Basic CRUD Operations

### Create a Scenario

```bash
curl -X POST http://localhost:8080/_uni/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "uuid": "scenario1",
    "path": "/api/users",
    "statusCode": 200,
    "contentType": "application/json",
    "data": "{\"message\":\"This is scenario 1\"}"
  }'
```

### List Scenarios

```bash
curl -X GET http://localhost:8080/_uni/scenarios
```

### Get a Specific Scenario

```bash
curl -X GET http://localhost:8080/_uni/scenarios/<scenario_id>
```

### Update a Scenario

```bash
curl -X PUT http://localhost:8080/_uni/scenarios/<scenario_id> \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/users",
    "statusCode": 201,
    "contentType": "application/json",
    "data": "{\"message\":\"Updated scenario\"}"
  }'
```

### Delete a Scenario

```bash
curl -X DELETE http://localhost:8080/_uni/scenarios/<scenario_id>
```

---

## For Developers

### Test Scenarios

| Scenario Name | Description | Expected Outcome |
|---------------|-------------|-----------------|
| Basic Operations | CRUD operations | Successful scenario creation, retrieval, updates, and deletion |
| Error Handling | Invalid IDs, missing data | Appropriate error codes and messages |

### Implementation Details

The scenario system consists of:

1. **ScenarioStorage**: Simple in-memory map with thread-safety
2. **ScenarioHandler**: HTTP handler for scenario management
3. **Direct Integration**: Used in the main application router

### Storage Interface

```go
type ScenarioStorage interface {
	Create(id string, scenario *model.Scenario) error
	Get(id string) (*model.Scenario, error)
	Update(id string, scenario *model.Scenario) error
	Delete(id string) error
	List() []*model.Scenario
}
```
