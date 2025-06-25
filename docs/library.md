# Go Library Usage

Embed Unimock directly in your Go application for testing, development, or as part of your service architecture.

## Installation

```bash
go get github.com/bmcszk/unimock/pkg
```

## Basic Usage

```go
import (
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
)

// Create configuration
mockConfig := &config.MockConfig{
    Sections: map[string]config.Section{
        "users": {
            PathPattern: "/api/users/*",
            BodyIDPaths: []string{"/id"},  // XPath-like syntax, not JSONPath
            ReturnBody:  true,
        },
    },
}

// Start embedded server
server, err := pkg.NewServer(
    pkg.WithPort(8080),
    pkg.WithMockConfig(mockConfig),
)
if err != nil {
    log.Fatal(err)
}

go server.ListenAndServe()
```

## Advanced Configuration

```go
import (
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
    "github.com/bmcszk/unimock/internal/model"
)

// Create comprehensive configuration
mockConfig := &config.MockConfig{
    Sections: map[string]config.Section{
        "users": {
            PathPattern:    "/api/users/*",
            BodyIDPaths:    []string{"/id", "/user/id"},
            HeaderIDName:   "X-User-ID", 
            ReturnBody:     true,
            Transformations: transformConfig, // Library-only transformations
        },
        "orders": {
            PathPattern:  "/api/orders/*",
            BodyIDPaths:  []string{"/order_id"},
            HeaderIDName: "X-Order-ID",
        },
    },
}

// Create server with scenarios
scenarios := []*model.Scenario{
    {
        UUID:        "user-not-found",
        Method:      "GET",
        Path:        "/api/users/999",
        StatusCode:  404,
        ContentType: "application/json",
        Data:        `{"error": "User not found"}`,
    },
}

server, err := pkg.NewServer(
    pkg.WithPort(8080),
    pkg.WithMockConfig(mockConfig),
    pkg.WithScenarios(scenarios),
    pkg.WithLogLevel("debug"),
)
```

## Data Transformations

Transform request and response data programmatically using Go functions. Transformations are **library-only** and cannot be configured via YAML.

### Creating Transformations

```go
import (
    "github.com/bmcszk/unimock/pkg/config"
    "github.com/bmcszk/unimock/internal/model"
)

// Create transformation config
transformConfig := config.NewTransformationConfig()

// Add request transformation (applied before storage)
transformConfig.AddRequestTransform(
    func(data *model.MockData) (*model.MockData, error) {
        // Modify request data before storing
        modifiedData := *data
        
        // Example: Add timestamp to request body
        var reqBody map[string]interface{}
        if err := json.Unmarshal(data.Body, &reqBody); err == nil {
            reqBody["created_at"] = time.Now().Format(time.RFC3339)
            if newBody, err := json.Marshal(reqBody); err == nil {
                modifiedData.Body = newBody
            }
        }
        
        return &modifiedData, nil
    })

// Add response transformation (applied after retrieval)
transformConfig.AddResponseTransform(
    func(data *model.MockData) (*model.MockData, error) {
        // Modify response data before returning
        modifiedData := *data
        
        // Example: Add computed field
        var respBody map[string]interface{}
        if err := json.Unmarshal(data.Body, &respBody); err == nil {
            respBody["server_time"] = time.Now().Unix()
            if newBody, err := json.Marshal(respBody); err == nil {
                modifiedData.Body = newBody
            }
        }
        
        return &modifiedData, nil
    })
```

### Using Transformations

```go
// Use in section configuration
section := config.Section{
    PathPattern:     "/api/users/*",
    BodyIDPaths:     []string{"/id"},
    Transformations: transformConfig,
}

mockConfig := &config.MockConfig{
    Sections: map[string]config.Section{
        "users": section,
    },
}
```

### Transformation Examples

#### Add Metadata to Responses

```go
transformConfig.AddResponseTransform(
    func(data *model.MockData) (*model.MockData, error) {
        modifiedData := *data
        
        var body map[string]interface{}
        if err := json.Unmarshal(data.Body, &body); err == nil {
            // Wrap response with metadata
            wrapped := map[string]interface{}{
                "data": body,
                "meta": map[string]interface{}{
                    "timestamp": time.Now().Format(time.RFC3339),
                    "version":   "1.0",
                    "source":    "unimock",
                },
            }
            
            if newBody, err := json.Marshal(wrapped); err == nil {
                modifiedData.Body = newBody
            }
        }
        
        return &modifiedData, nil
    })
```

#### Filter Sensitive Data

```go
transformConfig.AddResponseTransform(
    func(data *model.MockData) (*model.MockData, error) {
        modifiedData := *data
        
        var body map[string]interface{}
        if err := json.Unmarshal(data.Body, &body); err == nil {
            // Remove sensitive fields
            delete(body, "password")
            delete(body, "ssn")
            delete(body, "credit_card")
            
            if newBody, err := json.Marshal(body); err == nil {
                modifiedData.Body = newBody
            }
        }
        
        return &modifiedData, nil
    })
```

#### Transform Data Format

```go
transformConfig.AddRequestTransform(
    func(data *model.MockData) (*model.MockData, error) {
        modifiedData := *data
        
        // Convert snake_case to camelCase in request
        var body map[string]interface{}
        if err := json.Unmarshal(data.Body, &body); err == nil {
            converted := convertToCamelCase(body)
            if newBody, err := json.Marshal(converted); err == nil {
                modifiedData.Body = newBody
            }
        }
        
        return &modifiedData, nil
    })

func convertToCamelCase(data map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    for key, value := range data {
        camelKey := toCamelCase(key)
        result[camelKey] = value
    }
    return result
}
```

## Testing Integration

### Unit Testing with Embedded Server

```go
func TestWithEmbeddedUnimock(t *testing.T) {
    // Create test configuration
    mockConfig := &config.MockConfig{
        Sections: map[string]config.Section{
            "users": {
                PathPattern: "/api/users/*",
                BodyIDPaths: []string{"/id"},
                ReturnBody:  true,
            },
        },
    }
    
    // Start embedded server
    server, err := pkg.NewServer(
        pkg.WithPort(0), // Use random available port
        pkg.WithMockConfig(mockConfig),
    )
    require.NoError(t, err)
    
    // Start server in background
    go server.ListenAndServe()
    defer server.Shutdown(context.Background())
    
    // Get the actual port
    port := server.GetPort()
    baseURL := fmt.Sprintf("http://localhost:%d", port)
    
    // Use with HTTP client or Go client
    client, err := client.NewClient(baseURL)
    require.NoError(t, err)
    
    // Test operations
    user := map[string]interface{}{
        "id":   "test123",
        "name": "Test User",
    }
    
    resp, err := client.PostJSON(context.Background(), "/api/users", nil, user)
    require.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
    
    resp, err = client.Get(context.Background(), "/api/users/test123", nil)
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### Integration Testing with Scenarios

```go
func TestWithScenarios(t *testing.T) {
    scenarios := []*model.Scenario{
        {
            UUID:        "user-not-found",
            Method:      "GET",
            Path:        "/api/users/999",
            StatusCode:  404,
            ContentType: "application/json",
            Data:        `{"error": "User not found"}`,
        },
        {
            UUID:        "server-error",
            Method:      "POST",
            Path:        "/api/users",
            StatusCode:  500,
            ContentType: "application/json",
            Data:        `{"error": "Internal server error"}`,
        },
    }
    
    server, err := pkg.NewServer(
        pkg.WithPort(0),
        pkg.WithMockConfig(mockConfig),
        pkg.WithScenarios(scenarios),
    )
    require.NoError(t, err)
    
    go server.ListenAndServe()
    defer server.Shutdown(context.Background())
    
    // Test scenario responses
    baseURL := fmt.Sprintf("http://localhost:%d", server.GetPort())
    
    // Test 404 scenario
    resp, err := http.Get(baseURL + "/api/users/999")
    require.NoError(t, err)
    assert.Equal(t, 404, resp.StatusCode)
    
    // Test 500 scenario
    resp, err = http.Post(baseURL+"/api/users", "application/json", strings.NewReader(`{"name": "test"}`))
    require.NoError(t, err)
    assert.Equal(t, 500, resp.StatusCode)
}
```

### Test Helper Functions

```go
// Helper to create test server with common configuration
func createTestServer(t *testing.T, config *config.MockConfig) (*pkg.Server, string) {
    server, err := pkg.NewServer(
        pkg.WithPort(0),
        pkg.WithMockConfig(config),
        pkg.WithLogLevel("error"), // Reduce noise in tests
    )
    require.NoError(t, err)
    
    go server.ListenAndServe()
    t.Cleanup(func() {
        server.Shutdown(context.Background())
    })
    
    baseURL := fmt.Sprintf("http://localhost:%d", server.GetPort())
    return server, baseURL
}

// Helper to wait for server readiness
func waitForServer(t *testing.T, baseURL string, timeout time.Duration) {
    client := &http.Client{Timeout: 1 * time.Second}
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        resp, err := client.Get(baseURL + "/_uni/health")
        if err == nil && resp.StatusCode == 200 {
            resp.Body.Close()
            return
        }
        if resp != nil {
            resp.Body.Close()
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    t.Fatal("Server not ready within timeout")
}
```

## Server Configuration Options

### Available Options

```go
server, err := pkg.NewServer(
    pkg.WithPort(8080),                    // Server port (0 for random)
    pkg.WithMockConfig(mockConfig),        // Mock configuration
    pkg.WithScenarios(scenarios),          // Predefined scenarios
    pkg.WithLogLevel("debug"),             // Log level
    pkg.WithCORS(true),                    // Enable CORS
    pkg.WithHealthCheck("/health"),        // Custom health endpoint
    pkg.WithShutdownTimeout(30*time.Second), // Graceful shutdown timeout
)
```

### Configuration Loading

Load configuration from files:

```go
// Load from YAML file
mockConfig, err := config.LoadFromFile("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Load scenarios from YAML file  
scenarios, err := config.LoadScenariosFromFile("scenarios.yaml")
if err != nil {
    log.Fatal(err)
}

server, err := pkg.NewServer(
    pkg.WithMockConfig(mockConfig),
    pkg.WithScenarios(scenarios),
)
```

### Dynamic Configuration

Update configuration at runtime:

```go
// Get server instance
server, err := pkg.NewServer(pkg.WithPort(8080))

// Add new section dynamically
newSection := config.Section{
    PathPattern: "/api/products/*",
    BodyIDPaths: []string{"/id"},
}
server.AddSection("products", newSection)

// Add scenario dynamically
scenario := &model.Scenario{
    UUID:        "product-error",
    Method:      "GET", 
    Path:        "/api/products/error",
    StatusCode:  500,
    ContentType: "application/json",
    Data:        `{"error": "Product service unavailable"}`,
}
server.AddScenario(scenario)
```

## Production Usage

### Graceful Shutdown

```go
func main() {
    server, err := pkg.NewServer(
        pkg.WithPort(8080),
        pkg.WithMockConfig(mockConfig),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

### Health Monitoring

```go
// Custom health check
server.SetHealthCheck(func() error {
    // Custom health logic
    if someCondition {
        return errors.New("service unhealthy")
    }
    return nil
})

// Monitor server health
go func() {
    for {
        if err := server.Health(); err != nil {
            log.Printf("Health check failed: %v", err)
        }
        time.Sleep(30 * time.Second)
    }
}()
```

## Best Practices

### 1. Use Configuration Files

```go
// Load from file instead of hardcoding
mockConfig, err := config.LoadFromFile("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### 2. Handle Errors Gracefully

```go
// Always check transformation errors
transformConfig.AddRequestTransform(
    func(data *model.MockData) (*model.MockData, error) {
        // Safe transformation with error handling
        modifiedData := *data
        
        var body map[string]interface{}
        if err := json.Unmarshal(data.Body, &body); err != nil {
            // Log error but don't fail
            log.Printf("Transform error: %v", err)
            return &modifiedData, nil
        }
        
        // Transform logic here...
        
        return &modifiedData, nil
    })
```

### 3. Use Dependency Injection

```go
type TestSuite struct {
    server *pkg.Server
    client *client.Client
}

func (ts *TestSuite) SetupSuite() {
    server, err := pkg.NewServer(pkg.WithPort(0))
    require.NoError(ts.T(), err)
    
    go server.ListenAndServe()
    
    baseURL := fmt.Sprintf("http://localhost:%d", server.GetPort())
    client, err := client.NewClient(baseURL)
    require.NoError(ts.T(), err)
    
    ts.server = server
    ts.client = client
}

func (ts *TestSuite) TearDownSuite() {
    ts.server.Shutdown(context.Background())
}
```

### 4. Test Isolation

```go
func TestUserOperations(t *testing.T) {
    // Each test gets fresh server
    server, baseURL := createTestServer(t, mockConfig)
    client, _ := client.NewClient(baseURL)
    
    t.Run("CreateUser", func(t *testing.T) {
        // Test user creation
    })
    
    t.Run("GetUser", func(t *testing.T) {
        // Test user retrieval
    })
    
    // Server automatically cleaned up via t.Cleanup
}
```

The Go library provides full programmatic control over Unimock's behavior, making it ideal for integration testing, development environments, and embedded usage scenarios.