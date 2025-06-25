# Client Libraries

Use Unimock from your application code with these client libraries.

## Go Client

### Installation

```bash
go get github.com/bmcszk/unimock/pkg/client
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/bmcszk/unimock/pkg/client"
)

func main() {
    // Create client
    c := client.NewClient("http://localhost:8080")
    
    // Add test data
    user := map[string]interface{}{
        "id": "123",
        "name": "John Doe",
        "email": "john@example.com",
    }
    
    err := c.Post("/api/users", user)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve data
    var result map[string]interface{}
    err = c.Get("/api/users/123", &result)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("User: %+v\n", result)
}
```

### Advanced Usage

```go
// Custom headers
headers := map[string]string{
    "X-User-ID": "123",
    "Authorization": "Bearer token",
}
err := c.PostWithHeaders("/api/users", user, headers)

// Update data
user["name"] = "Jane Doe"
err = c.Put("/api/users/123", user)

// Delete data
err = c.Delete("/api/users/123")

// Raw HTTP response
resp, err := c.GetRaw("/api/users/123")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// Check status
if resp.StatusCode != 200 {
    fmt.Printf("Error: %d\n", resp.StatusCode)
}
```

### Error Handling

```go
// Check for specific errors
err := c.Get("/api/users/999", &result)
if err != nil {
    if client.IsNotFound(err) {
        fmt.Println("User not found")
    } else if client.IsServerError(err) {
        fmt.Println("Server error")
    } else {
        log.Fatal(err)
    }
}
```

## HTTP/REST Examples

### cURL

```bash
# Add data
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "John Doe"}'

# Get data
curl http://localhost:8080/api/users/123

# Update data
curl -X PUT http://localhost:8080/api/users/123 \
  -H "Content-Type: application/json" \
  -d '{"id": "123", "name": "Jane Doe"}'

# Delete data
curl -X DELETE http://localhost:8080/api/users/123
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

class UnimockClient {
    constructor(baseURL = 'http://localhost:8080') {
        this.client = axios.create({ baseURL });
    }
    
    async post(path, data) {
        const response = await this.client.post(path, data);
        return response.data;
    }
    
    async get(path) {
        const response = await this.client.get(path);
        return response.data;
    }
    
    async put(path, data) {
        const response = await this.client.put(path, data);
        return response.data;
    }
    
    async delete(path) {
        await this.client.delete(path);
    }
}

// Usage
const unimock = new UnimockClient();

async function example() {
    // Add user
    await unimock.post('/api/users', {
        id: '123',
        name: 'John Doe',
        email: 'john@example.com'
    });
    
    // Get user
    const user = await unimock.get('/api/users/123');
    console.log('User:', user);
    
    // Update user
    await unimock.put('/api/users/123', {
        id: '123',
        name: 'Jane Doe',
        email: 'jane@example.com'
    });
    
    // Delete user
    await unimock.delete('/api/users/123');
}

example().catch(console.error);
```

### Python

```python
import requests
import json

class UnimockClient:
    def __init__(self, base_url='http://localhost:8080'):
        self.base_url = base_url
        self.session = requests.Session()
    
    def post(self, path, data):
        url = f"{self.base_url}{path}"
        response = self.session.post(url, json=data)
        response.raise_for_status()
        return response.json() if response.content else None
    
    def get(self, path):
        url = f"{self.base_url}{path}"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()
    
    def put(self, path, data):
        url = f"{self.base_url}{path}"
        response = self.session.put(url, json=data)
        response.raise_for_status()
        return response.json() if response.content else None
    
    def delete(self, path):
        url = f"{self.base_url}{path}"
        response = self.session.delete(url)
        response.raise_for_status()

# Usage
client = UnimockClient()

# Add user
client.post('/api/users', {
    'id': '123',
    'name': 'John Doe',
    'email': 'john@example.com'
})

# Get user
user = client.get('/api/users/123')
print(f"User: {user}")

# Update user
client.put('/api/users/123', {
    'id': '123',
    'name': 'Jane Doe',
    'email': 'jane@example.com'
})

# Delete user
client.delete('/api/users/123')
```

### Java

```java
import com.fasterxml.jackson.databind.ObjectMapper;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.URI;
import java.util.Map;

public class UnimockClient {
    private final HttpClient client;
    private final String baseUrl;
    private final ObjectMapper mapper;
    
    public UnimockClient(String baseUrl) {
        this.client = HttpClient.newHttpClient();
        this.baseUrl = baseUrl;
        this.mapper = new ObjectMapper();
    }
    
    public void post(String path, Object data) throws Exception {
        String json = mapper.writeValueAsString(data);
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + path))
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(json))
            .build();
        
        client.send(request, HttpResponse.BodyHandlers.ofString());
    }
    
    public <T> T get(String path, Class<T> responseType) throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + path))
            .GET()
            .build();
        
        HttpResponse<String> response = client.send(request, 
            HttpResponse.BodyHandlers.ofString());
        
        return mapper.readValue(response.body(), responseType);
    }
}

// Usage
UnimockClient client = new UnimockClient("http://localhost:8080");

// Add user
Map<String, Object> user = Map.of(
    "id", "123",
    "name", "John Doe",
    "email", "john@example.com"
);
client.post("/api/users", user);

// Get user
Map userResult = client.get("/api/users/123", Map.class);
System.out.println("User: " + userResult);
```

## Testing Integration

### Go Testing

```go
package main

import (
    "testing"
    "github.com/bmcszk/unimock/pkg/client"
)

func TestUserAPI(t *testing.T) {
    // Setup
    c := client.NewClient("http://localhost:8080")
    
    // Test data
    user := map[string]interface{}{
        "id": "test123",
        "name": "Test User",
        "email": "test@example.com",
    }
    
    // Create user
    err := c.Post("/api/users", user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    // Verify user exists
    var result map[string]interface{}
    err = c.Get("/api/users/test123", &result)
    if err != nil {
        t.Fatalf("Failed to get user: %v", err)
    }
    
    // Check fields
    if result["name"] != "Test User" {
        t.Errorf("Expected name 'Test User', got %v", result["name"])
    }
    
    // Cleanup
    err = c.Delete("/api/users/test123")
    if err != nil {
        t.Fatalf("Failed to delete user: %v", err)
    }
}
```

### Jest (JavaScript)

```javascript
const UnimockClient = require('./unimock-client');

describe('User API', () => {
    let client;
    
    beforeEach(() => {
        client = new UnimockClient('http://localhost:8080');
    });
    
    test('should create and retrieve user', async () => {
        // Create user
        const userData = {
            id: 'test123',
            name: 'Test User',
            email: 'test@example.com'
        };
        
        await client.post('/api/users', userData);
        
        // Retrieve user
        const user = await client.get('/api/users/test123');
        
        expect(user.name).toBe('Test User');
        expect(user.email).toBe('test@example.com');
        
        // Cleanup
        await client.delete('/api/users/test123');
    });
    
    test('should handle 404 for non-existent user', async () => {
        await expect(client.get('/api/users/nonexistent'))
            .rejects.toThrow();
    });
});
```

### pytest (Python)

```python
import pytest
from unimock_client import UnimockClient

@pytest.fixture
def client():
    return UnimockClient('http://localhost:8080')

def test_user_crud(client):
    user_data = {
        'id': 'test123',
        'name': 'Test User',
        'email': 'test@example.com'
    }
    
    # Create
    client.post('/api/users', user_data)
    
    # Read
    user = client.get('/api/users/test123')
    assert user['name'] == 'Test User'
    assert user['email'] == 'test@example.com'
    
    # Update
    user_data['name'] = 'Updated User'
    client.put('/api/users/test123', user_data)
    
    updated_user = client.get('/api/users/test123')
    assert updated_user['name'] == 'Updated User'
    
    # Delete
    client.delete('/api/users/test123')
    
    # Verify deletion
    with pytest.raises(requests.exceptions.HTTPError):
        client.get('/api/users/test123')
```

## Best Practices

### 1. Use Environment Variables

```go
// Don't hardcode URLs
baseURL := os.Getenv("UNIMOCK_URL")
if baseURL == "" {
    baseURL = "http://localhost:8080"  // fallback
}
client := client.NewClient(baseURL)
```

### 2. Clean Up Test Data

```go
func TestWithCleanup(t *testing.T) {
    client := client.NewClient("http://localhost:8080")
    
    // Clean up after test
    defer func() {
        client.Delete("/api/users/test123")
    }()
    
    // Your test code here
}
```

### 3. Use Unique Test IDs

```go
// Generate unique IDs to avoid conflicts
testID := fmt.Sprintf("test-%d", time.Now().UnixNano())
user := map[string]interface{}{
    "id": testID,
    "name": "Test User",
}
```

### 4. Check Health Before Tests

```go
func TestMain(m *testing.M) {
    // Wait for Unimock to be ready
    client := client.NewClient("http://localhost:8080")
    
    for i := 0; i < 30; i++ {
        if err := client.Health(); err == nil {
            break
        }
        time.Sleep(time.Second)
    }
    
    // Run tests
    code := m.Run()
    os.Exit(code)
}
```

This ensures Unimock is running before your tests start.