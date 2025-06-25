# Go Client Library

The Unimock Go client provides a comprehensive interface for interacting with Unimock servers, including both HTTP operations and scenario management.

## Installation

```bash
go get github.com/bmcszk/unimock/pkg/client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/bmcszk/unimock/pkg/client"
)

func main() {
    // Create client (uses localhost:8080 if URL is empty)
    c, err := client.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store data via POST
    headers := map[string]string{"Content-Type": "application/json"}
    data := `{"id": "123", "name": "John Doe", "email": "john@example.com"}`
    
    resp, err := c.Post(ctx, "/api/users", headers, []byte(data))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created: %d\n", resp.StatusCode)
    
    // Retrieve data via GET
    resp, err = c.Get(ctx, "/api/users/123", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Retrieved: %s\n", string(resp.Body))
}
```

## Client Creation

```go
// Connect to localhost:8080 (default)
client, err := client.NewClient("")

// Connect to custom URL
client, err := client.NewClient("http://localhost:9090")
client, err := client.NewClient("https://unimock.example.com")

// The client includes a 10-second timeout HTTP client by default
```

## HTTP Methods

The client supports all standard HTTP methods with consistent signatures:

### Basic HTTP Operations

```go
ctx := context.Background()
headers := map[string]string{
    "Content-Type": "application/json",
    "X-Custom-Header": "value",
}

// GET request
resp, err := client.Get(ctx, "/api/users/123", headers)

// HEAD request
resp, err := client.Head(ctx, "/api/users/123", headers)

// POST request
body := []byte(`{"name": "John Doe"}`)
resp, err := client.Post(ctx, "/api/users", headers, body)

// PUT request
body := []byte(`{"id": "123", "name": "Jane Doe"}`)
resp, err := client.Put(ctx, "/api/users/123", headers, body)

// PATCH request
body := []byte(`{"name": "Jane Smith"}`)
resp, err := client.Patch(ctx, "/api/users/123", headers, body)

// DELETE request
resp, err := client.Delete(ctx, "/api/users/123", headers)

// OPTIONS request
resp, err := client.Options(ctx, "/api/users", headers)
```

### JSON Convenience Methods

For JSON requests, use the convenience methods that handle serialization:

```go
// POST with automatic JSON marshaling
user := map[string]interface{}{
    "id":    "123",
    "name":  "John Doe",
    "email": "john@example.com",
}

resp, err := client.PostJSON(ctx, "/api/users", headers, user)

// PUT with automatic JSON marshaling
user["name"] = "Jane Doe"
resp, err := client.PutJSON(ctx, "/api/users/123", headers, user)

// PATCH with automatic JSON marshaling
updates := map[string]string{"name": "Jane Smith"}
resp, err := client.PatchJSON(ctx, "/api/users/123", headers, updates)
```

## Response Handling

All methods return a `Response` struct:

```go
type Response struct {
    StatusCode int         // HTTP status code (200, 404, etc.)
    Headers    http.Header // Response headers
    Body       []byte      // Response body content
}
```

Example response handling:

```go
resp, err := client.Get(ctx, "/api/users/123", nil)
if err != nil {
    log.Fatal(err)
}

// Check status code
switch resp.StatusCode {
case 200:
    fmt.Printf("Success: %s\n", string(resp.Body))
case 404:
    fmt.Println("User not found")
case 500:
    fmt.Println("Server error")
default:
    fmt.Printf("Unexpected status: %d\n", resp.StatusCode)
}

// Access headers
contentType := resp.Headers.Get("Content-Type")
fmt.Printf("Content-Type: %s\n", contentType)

// Parse JSON response
var user map[string]interface{}
if err := json.Unmarshal(resp.Body, &user); err != nil {
    log.Fatal(err)
}
```

## Scenario Management

The client provides full CRUD operations for managing scenarios:

### Create Scenario

```go
import "github.com/bmcszk/unimock/internal/model"

scenario := &model.Scenario{
    UUID:        "user-not-found", // Optional, auto-generated if empty
    Method:      "GET",
    Path:        "/api/users/999",
    StatusCode:  404,
    ContentType: "application/json",
    Data:        `{"error": "User not found", "code": "USER_NOT_FOUND"}`,
}

err := client.CreateScenario(ctx, scenario)
if err != nil {
    log.Fatal(err)
}
```

### Get Scenario

```go
// Get specific scenario by UUID
scenario, err := client.GetScenario(ctx, "user-not-found")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Scenario: %s %s -> %d\n", scenario.Method, scenario.Path, scenario.StatusCode)
```

### List All Scenarios

```go
scenarios, err := client.ListScenarios(ctx)
if err != nil {
    log.Fatal(err)
}

for _, scenario := range scenarios {
    fmt.Printf("- %s: %s %s\n", scenario.UUID, scenario.Method, scenario.Path)
}
```

### Update Scenario

```go
// Update existing scenario
scenario.StatusCode = 410
scenario.Data = `{"error": "User permanently deleted"}`

err := client.UpdateScenario(ctx, "user-not-found", scenario)
if err != nil {
    log.Fatal(err)
}
```

### Delete Scenario

```go
err := client.DeleteScenario(ctx, "user-not-found")
if err != nil {
    log.Fatal(err)
}
```

## Health Check

Check if the Unimock server is healthy:

```go
err := client.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
} else {
    fmt.Println("Server is healthy")
}
```

## Error Handling

The client returns detailed errors for different failure conditions:

```go
resp, err := client.Get(ctx, "/api/users/999", nil)
if err != nil {
    // Network or client errors
    log.Printf("Request failed: %v", err)
    return
}

// Check HTTP status codes
switch resp.StatusCode {
case 200:
    // Success
case 404:
    fmt.Println("Resource not found")
case 409:
    fmt.Println("Conflict - resource already exists")
case 500:
    fmt.Println("Server error")
default:
    fmt.Printf("Unexpected status: %d\n", resp.StatusCode)
}
```

## URL Handling

The client supports both relative paths and absolute URLs:

```go
// Relative paths (recommended)
resp, err := client.Get(ctx, "/api/users/123", nil)

// Absolute URLs (overrides base URL)
resp, err := client.Get(ctx, "https://api.example.com/users/123", nil)

// URLs are automatically cleaned and joined properly
resp, err := client.Get(ctx, "api/users/123", nil)    // Leading slash added
resp, err := client.Get(ctx, "/api/users/123/", nil)  // Trailing slash preserved
```

## Testing Integration

### Unit Testing

```go
func TestUserAPI(t *testing.T) {
    client, err := client.NewClient("http://localhost:8080")
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Test data
    user := map[string]interface{}{
        "id":    "test123",
        "name":  "Test User",
        "email": "test@example.com",
    }
    
    // Create user
    resp, err := client.PostJSON(ctx, "/api/users", nil, user)
    require.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
    
    // Verify user exists
    resp, err = client.Get(ctx, "/api/users/test123", nil)
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Parse response
    var result map[string]interface{}
    err = json.Unmarshal(resp.Body, &result)
    require.NoError(t, err)
    assert.Equal(t, "Test User", result["name"])
    
    // Cleanup
    resp, err = client.Delete(ctx, "/api/users/test123", nil)
    require.NoError(t, err)
}
```

### Integration Testing with Scenarios

```go
func TestUserNotFoundScenario(t *testing.T) {
    client, err := client.NewClient("http://localhost:8080")
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Create scenario for user not found
    scenario := &model.Scenario{
        Method:      "GET",
        Path:        "/api/users/999",
        StatusCode:  404,
        ContentType: "application/json",
        Data:        `{"error": "User not found"}`,
    }
    
    err = client.CreateScenario(ctx, scenario)
    require.NoError(t, err)
    
    // Cleanup scenario after test
    defer func() {
        client.DeleteScenario(ctx, scenario.UUID)
    }()
    
    // Test the scenario
    resp, err := client.Get(ctx, "/api/users/999", nil)
    require.NoError(t, err)
    assert.Equal(t, 404, resp.StatusCode)
    
    var errorResp map[string]string
    err = json.Unmarshal(resp.Body, &errorResp)
    require.NoError(t, err)
    assert.Equal(t, "User not found", errorResp["error"])
}
```

### Test Helpers

```go
// Helper function for creating test scenarios
func createTestScenario(t *testing.T, client *client.Client, method, path string, statusCode int, data string) string {
    scenario := &model.Scenario{
        Method:      method,
        Path:        path,
        StatusCode:  statusCode,
        ContentType: "application/json",
        Data:        data,
    }
    
    err := client.CreateScenario(context.Background(), scenario)
    require.NoError(t, err)
    
    // Return UUID for cleanup
    return scenario.UUID
}

// Helper function for cleanup
func cleanupScenario(t *testing.T, client *client.Client, uuid string) {
    err := client.DeleteScenario(context.Background(), uuid)
    require.NoError(t, err)
}
```

## Best Practices

### 1. Use Context for Timeouts

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.Get(ctx, "/api/users/123", nil)
```

### 2. Clean Up Test Resources

```go
func TestUserCRUD(t *testing.T) {
    client, _ := client.NewClient("http://localhost:8080")
    ctx := context.Background()
    
    // Create test user
    user := map[string]string{"id": "test123", "name": "Test User"}
    client.PostJSON(ctx, "/api/users", nil, user)
    
    // Ensure cleanup
    defer func() {
        client.Delete(ctx, "/api/users/test123", nil)
    }()
    
    // Your test logic here...
}
```

### 3. Use Environment Variables for URLs

```go
func NewTestClient() (*client.Client, error) {
    url := os.Getenv("UNIMOCK_URL")
    if url == "" {
        url = "http://localhost:8080" // fallback
    }
    return client.NewClient(url)
}
```

### 4. Check Server Health Before Tests

```go
func TestMain(m *testing.M) {
    client, err := client.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for server to be ready
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    for {
        if err := client.HealthCheck(ctx); err == nil {
            break
        }
        
        select {
        case <-ctx.Done():
            log.Fatal("Server not ready within timeout")
        case <-time.After(time.Second):
            // Continue waiting
        }
    }
    
    // Run tests
    code := m.Run()
    os.Exit(code)
}
```

### 5. Use Unique Test IDs

```go
// Generate unique test IDs to avoid conflicts
testID := fmt.Sprintf("test-%d", time.Now().UnixNano())
user := map[string]string{
    "id":   testID,
    "name": "Test User",
}
```

## Advanced Usage

### Custom Headers for All Requests

```go
// Create headers map that you reuse
commonHeaders := map[string]string{
    "Authorization": "Bearer " + token,
    "X-API-Version": "v1",
}

// Use in requests
resp, err := client.Get(ctx, "/api/users", commonHeaders)
```

### Working with Different Content Types

```go
// XML request
xmlData := `<?xml version="1.0"?><user><name>John</name></user>`
headers := map[string]string{"Content-Type": "application/xml"}
resp, err := client.Post(ctx, "/api/users", headers, []byte(xmlData))

// Plain text request
textData := "Simple text data"
headers = map[string]string{"Content-Type": "text/plain"}
resp, err = client.Post(ctx, "/api/logs", headers, []byte(textData))

// Binary data
binaryData, _ := ioutil.ReadFile("image.jpg")
headers = map[string]string{"Content-Type": "image/jpeg"}
resp, err = client.Post(ctx, "/api/uploads", headers, binaryData)
```

The Go client provides a complete interface for both basic HTTP operations and advanced scenario management, making it easy to integrate Unimock into your Go applications and tests.