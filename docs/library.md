# Go Library Usage

Embed Unimock directly in your Go application for testing, development, or as part of your service architecture.

## Installation

```bash
go get github.com/bmcszk/unimock/pkg
```

## Basic Usage

```go
import (
    "log"
    "net/http"
    
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
)

func main() {
    // Create server configuration
    serverConfig := &config.ServerConfig{
        Port:       "8080",
        LogLevel:   "info",
        ConfigPath: "config.yaml",
    }

    // Create mock configuration
    uniConfig := &config.UniConfig{
        Sections: map[string]config.Section{
            "users": {
                PathPattern:   "/api/users/*",
                BodyIDPaths:   []string{"/id"},  // XPath-like syntax for JSON/XML
                HeaderIDNames: []string{"X-User-ID"},
                ReturnBody:    true,
            },
        },
    }

    // Create server
    server, err := pkg.NewServer(serverConfig, uniConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Start server
    log.Printf("Starting server on port %s", serverConfig.Port)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

## Advanced Configuration

```go
import (
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
)

func main() {
    // Load server config from environment variables
    // UNIMOCK_PORT, UNIMOCK_LOG_LEVEL, UNIMOCK_CONFIG
    serverConfig := config.FromEnv()

    // Create comprehensive mock configuration
    uniConfig := &config.UniConfig{
        Sections: map[string]config.Section{
            "users": {
                PathPattern:     "/api/users/*",
                BodyIDPaths:     []string{"/id", "/user/id"},
                HeaderIDNames:   []string{"X-User-ID"}, 
                ReturnBody:      true,
                Transformations: createUserTransformations(), // Library-only
            },
            "orders": {
                PathPattern:   "/api/orders/*",
                BodyIDPaths:   []string{"/order_id"},
                HeaderIDNames: []string{"X-Order-ID"},
                ReturnBody:    false,
            },
        },
    }

    // Create server with configuration
    server, err := pkg.NewServer(serverConfig, uniConfig)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Server starting on port %s", serverConfig.Port)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

## Data Transformations

Transform request and response data programmatically using Go functions. Transformations are **library-only** and cannot be configured via YAML.

### Creating Transformations

```go
import (
    "encoding/json"
    "time"
    
    "github.com/bmcszk/unimock/pkg/config"
    "github.com/bmcszk/unimock/pkg/model"
)

// Create transformation configuration
func createUserTransformations() *config.TransformationConfig {
    transformConfig := config.NewTransformationConfig()

    // Add request transformation (applied before storage)
    transformConfig.AddRequestTransform(
        func(data model.UniData) (model.UniData, error) {
            // Modify request data before storing
            modifiedData := data
            
            // Example: Add timestamp to request body
            var reqBody map[string]interface{}
            if err := json.Unmarshal(data.Body, &reqBody); err == nil {
                reqBody["created_at"] = time.Now().Format(time.RFC3339)
                if newBody, err := json.Marshal(reqBody); err == nil {
                    modifiedData.Body = newBody
                }
            }
            
            return modifiedData, nil
        })

    // Add response transformation (applied after retrieval)
    transformConfig.AddResponseTransform(
        func(data model.UniData) (model.UniData, error) {
            // Modify response data before returning
            modifiedData := data
            
            // Example: Add computed field
            var respBody map[string]interface{}
            if err := json.Unmarshal(data.Body, &respBody); err == nil {
                respBody["server_time"] = time.Now().Unix()
                if newBody, err := json.Marshal(respBody); err == nil {
                    modifiedData.Body = newBody
                }
            }
            
            return modifiedData, nil
        })

    return transformConfig
}
```

### Using Transformations in Sections

```go
// Create section with transformations
section := config.Section{
    PathPattern:     "/api/users/*",
    BodyIDPaths:     []string{"/id"},
    ReturnBody:      true,
    Transformations: createUserTransformations(),
}

uniConfig := &config.UniConfig{
    Sections: map[string]config.Section{
        "users": section,
    },
}
```

### Transformation Examples

#### Add Metadata to Responses

```go
func addMetadataTransform() *config.TransformationConfig {
    transformConfig := config.NewTransformationConfig()
    
    transformConfig.AddResponseTransform(
        func(data model.UniData) (model.UniData, error) {
            modifiedData := data
            
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
            
            return modifiedData, nil
        })
    
    return transformConfig
}
```

#### Filter Sensitive Data

```go
func filterSensitiveDataTransform() *config.TransformationConfig {
    transformConfig := config.NewTransformationConfig()
    
    transformConfig.AddResponseTransform(
        func(data model.UniData) (model.UniData, error) {
            modifiedData := data
            
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
            
            return modifiedData, nil
        })
    
    return transformConfig
}
```

## Configuration Loading

### From YAML Files

```go
import (
    "github.com/bmcszk/unimock/pkg/config"
)

func loadFromYAML() (*config.UniConfig, error) {
    // Load mock configuration from YAML file
    return config.LoadFromYAML("config.yaml")
}

func loadUnifiedFromYAML() (*config.UnifiedConfig, error) {
    // Load unified configuration (sections + scenarios) from YAML file  
    return config.LoadUnifiedFromYAML("config.yaml")
}
```

### From Environment Variables

```go
// Load server configuration from environment variables
serverConfig := config.FromEnv()

// Environment variables:
// UNIMOCK_PORT - Server port (default: "8080")
// UNIMOCK_LOG_LEVEL - Log level (default: "info")
// UNIMOCK_CONFIG - Config file path (default: "config.yaml")
```

## Testing Integration

### Unit Testing with Embedded Server

```go
import (
    "context"
    "fmt"
    "net/http"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/bmcszk/unimock/pkg"
    "github.com/bmcszk/unimock/pkg/config"
)

func TestWithEmbeddedUnimock(t *testing.T) {
    // Create test configuration
    serverConfig := &config.ServerConfig{
        Port:       "0", // Use random available port
        LogLevel:   "error", // Reduce noise in tests
        ConfigPath: "config.yaml",
    }
    
    uniConfig := &config.UniConfig{
        Sections: map[string]config.Section{
            "users": {
                PathPattern:   "/api/users/*",
                BodyIDPaths:   []string{"/id"},
                HeaderIDNames: []string{"X-User-ID"},
                ReturnBody:    true,
            },
        },
    }
    
    // Start embedded server
    server, err := pkg.NewServer(serverConfig, uniConfig)
    require.NoError(t, err)
    
    // Start server in background
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            t.Errorf("Server error: %v", err)
        }
    }()
    
    // Cleanup
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        server.Shutdown(ctx)
    }()
    
    // Wait for server to start
    time.Sleep(100 * time.Millisecond)
    
    // Get server address from actual listener
    baseURL := fmt.Sprintf("http://%s", server.Addr)
    
    // Test operations using standard HTTP client
    client := &http.Client{Timeout: 5 * time.Second}
    
    // Create user
    resp, err := client.Post(baseURL+"/api/users", "application/json", 
        strings.NewReader(`{"id": "test123", "name": "Test User"}`))
    require.NoError(t, err)
    defer resp.Body.Close()
    assert.Equal(t, 201, resp.StatusCode)
    
    // Get user
    resp, err = client.Get(baseURL + "/api/users/test123")
    require.NoError(t, err)
    defer resp.Body.Close()
    assert.Equal(t, 200, resp.StatusCode)
}
```

### Integration Testing with File-based Configuration

```go
func TestWithConfigFiles(t *testing.T) {
    // Create temporary config file
    configFile := createTempConfig(t, `
sections:
  users:
    path_pattern: "/api/users/*"
    body_id_paths:
      - "/id"
    return_body: true
`)
    defer os.Remove(configFile)
    
    // Create server config pointing to file
    serverConfig := &config.ServerConfig{
        Port:       "0",
        LogLevel:   "error",
        ConfigPath: configFile,
    }
    
    // Load mock config from file
    uniConfig, err := config.LoadFromYAML(serverConfig.ConfigPath)
    require.NoError(t, err)
    
    // Create and test server
    server, err := pkg.NewServer(serverConfig, uniConfig)
    require.NoError(t, err)
    
    // ... test operations
}

func createTempConfig(t *testing.T, content string) string {
    file, err := os.CreateTemp("", "unimock-test-*.yaml")
    require.NoError(t, err)
    
    _, err = file.WriteString(content)
    require.NoError(t, err)
    
    err = file.Close()
    require.NoError(t, err)
    
    return file.Name()
}
```

## Server Configuration Options

### Available Configuration

```go
// Server configuration
serverConfig := &config.ServerConfig{
    Port:       "8080",        // Server port
    LogLevel:   "info",        // Log level: debug, info, warn, error
    ConfigPath: "config.yaml", // Path to mock config file
}

// Mock configuration  
uniConfig := &config.UniConfig{
    Sections: map[string]config.Section{
        "section_name": {
            PathPattern:     "/api/path/*",           // URL pattern
            BodyIDPaths:     []string{"/id"},         // JSON/XML ID paths
            HeaderIDNames:   []string{"X-Resource-ID"}, // Header ID names
            ReturnBody:      true,                   // Return request body on GET
            StrictPath:      false,                  // Strict path matching
            Transformations: transformationConfig,    // Library-only transformations
        },
    },
}
```

### Default Values

```go
// Create with defaults
serverConfig := config.NewDefaultServerConfig()
// Defaults: Port="8080", LogLevel="info", ConfigPath="config.yaml"

// Load from environment variables
serverConfig := config.FromEnv()
// Reads: UNIMOCK_PORT, UNIMOCK_LOG_LEVEL, UNIMOCK_CONFIG
```

## Production Usage

### Graceful Shutdown

```go
import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    server, err := pkg.NewServer(serverConfig, uniConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server in background
    go func() {
        log.Printf("Starting server on %s", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server error:", err)
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

### Error Handling

```go
// Handle configuration errors
server, err := pkg.NewServer(serverConfig, mockConfig)
if err != nil {
    if configErr, ok := err.(*pkg.ConfigError); ok {
        log.Printf("Configuration error: %s", configErr.Message)
        // Handle configuration-specific errors
    } else {
        log.Printf("Server creation failed: %v", err)
    }
    return
}

// Handle transformation errors (transformations return 500 on error)
transformConfig.AddRequestTransform(
    func(data model.UniData) (model.UniData, error) {
        // Safe transformation with error handling
        modifiedData := data
        
        var body map[string]interface{}
        if err := json.Unmarshal(data.Body, &body); err != nil {
            // Return error - will result in HTTP 500
            return model.UniData{}, fmt.Errorf("invalid JSON in request: %w", err)
        }
        
        // Transform logic here...
        
        return modifiedData, nil
    })
```

## Best Practices

### 1. Use Configuration Files

```go
// Load from file instead of hardcoding
uniConfig, err := config.LoadFromYAML("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### 2. Handle Errors Gracefully

```go
// Always check transformation errors
transformConfig.AddRequestTransform(
    func(data model.UniData) (model.UniData, error) {
        // Safe transformation with error handling
        modifiedData := data
        
        var body map[string]interface{}
        if err := json.Unmarshal(data.Body, &body); err != nil {
            // Log error but don't fail for malformed JSON
            log.Printf("Transform error: %v", err)
            return modifiedData, nil
        }
        
        // Transform logic here...
        
        return modifiedData, nil
    })
```

### 3. Use Environment Variables for Deployment

```go
// Production-ready configuration
serverConfig := config.FromEnv()

// Allows deployment-time configuration via:
// UNIMOCK_PORT=8080
// UNIMOCK_LOG_LEVEL=info
// UNIMOCK_CONFIG=/etc/unimock/config.yaml
```

### 4. Test Isolation

```go
func TestUserOperations(t *testing.T) {
    // Each test gets fresh server
    server, baseURL := createTestServer(t, uniConfig)
    defer shutdownServer(server)
    
    t.Run("CreateUser", func(t *testing.T) {
        // Test user creation
    })
    
    t.Run("GetUser", func(t *testing.T) {
        // Test user retrieval
    })
}

func createTestServer(t *testing.T, uniConfig *config.UniConfig) (*http.Server, string) {
    serverConfig := &config.ServerConfig{
        Port:       "0", 
        LogLevel:   "error",
        ConfigPath: "config.yaml",
    }
    server, err := pkg.NewServer(serverConfig, uniConfig)
    require.NoError(t, err)
    
    go server.ListenAndServe()
    time.Sleep(50 * time.Millisecond) // Wait for startup
    
    return server, fmt.Sprintf("http://%s", server.Addr)
}
```

The Go library provides programmatic control over Unimock's behavior, making it ideal for integration testing, development environments, and embedded usage scenarios.