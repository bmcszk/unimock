# Testing Guidelines

This document defines ALL testing rules and patterns for the BLACKSWAN project.

## Test Types

### 1. Unit Tests
- **Location**: Same package as code being tested
- **Package Name**: `package <name>_test` (e.g., `package handlers_test`)
- **Dependencies**: Mock ALL dependencies using `testify/mock` library
- **Scope**: Test ONLY public methods and interfaces
- **Speed**: Fast (<10ms per test)
- **Isolation**: No external dependencies (databases, networks, files)
- **Build Tags**: **DO NOT USE** - directory-based separation only

#### Unit Test Isolation Principles (CRITICAL)
**Unit tests must be completely isolated from external systems:**

**❌ Anti-Pattern: External Dependencies in Unit Tests**
```go
type HealthHandler struct {
    backendURL string
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // BAD: Makes real HTTP call during unit test
    resp, err := http.Get(h.backendURL + "/health")
    // This will fail in CI where no services are running
}
```

**✅ Correct Pattern: Dependency Injection for Testability**
```go
type HealthChecker interface {
    IsHealthy(ctx context.Context) bool
}

type HealthHandler struct {
    healthChecker HealthChecker
}

// Production constructor
func NewHealthHandler(backendURL string) *HealthHandler {
    return &HealthHandler{
        healthChecker: NewBackendHealthChecker(backendURL),
    }
}

// Test constructor with dependency injection
func NewHealthHandlerWithChecker(checker HealthChecker) *HealthHandler {
    return &HealthHandler{healthChecker: checker}
}
```

**Mock Implementation for Testing:**
```go
type MockHealthChecker struct {
    healthy bool
}

func (m *MockHealthChecker) IsHealthy(_ context.Context) bool {
    return m.healthy
}

func TestHealthHandler_Healthy(t *testing.T) {
    // given
    mockChecker := &MockHealthChecker{healthy: true}
    handler := NewHealthHandlerWithChecker(mockChecker)

    // when
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)

    // then
    assert.Equal(t, http.StatusOK, w.Code)
}
```

#### Mandatory Unit Testing Patterns
1. **Constructor Overloading**: Provide both production and test constructors
2. **Interface Segregation**: Define minimal interfaces for dependencies
3. **Test Naming Convention**: `TestComponent_Scenario_ExpectedResult`
4. **CI Environment Assumptions**: Tests must pass without any external services

#### Unit Test Structure
**ALL tests MUST use Given/When/Then structure for readability:**

```go
package handlers_test

import (
    "testing"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/assert"
    "github.com/bmcszk/blackswan/services/handlers"
)

func TestHandler_PublicMethod(t *testing.T) {
    // given
    mockDep := &MockDependency{}
    mockDep.On("Method", mock.Anything).Return(expectedResult, nil)
    handler := handlers.NewHandler(mockDep)
    
    // when
    result, err := handler.PublicMethod()
    
    // then
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockDep.AssertExpectations(t)
}
```

### 2. Integration Tests
- **Location**: `test/integration/` directory
- **Package Name**: `package integration_test`
- **Dependencies**: Real external services using `testcontainers` library
- **Scope**: Test component interactions with real dependencies
- **Speed**: Moderate (100ms-1s per test)
- **External Services**: Use testcontainers for databases, message queues, etc.
- **Build Tags**: **DO NOT USE** - directory-based separation only

#### Integration Test Structure
**ALL tests MUST use Given/When/Then structure for readability:**

```go
package integration_test

import (
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/bmcszk/blackswan/services/service"
)

func TestService_DatabaseIntegration(t *testing.T) {
    // given
    dbContainer := startPostgresContainer(t)
    defer dbContainer.Terminate(context.Background())
    service := service.NewService(dbContainer.ConnectionString())
    
    // when
    result, err := service.ProcessData()
    
    // then
    assert.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

### 3. E2E Tests
- **Location**: `test/e2e/` directory  
- **Package Name**: `package e2e_test`
- **Dependencies**: Full system stack using `docker-compose.test.yml`
- **Scope**: Test complete user workflows across all services
- **Speed**: Slow (1s+ per test)
- **Environment**: Tests run OUTSIDE containers, connecting to Docker Compose services
- **Execution**: Tests connect to services running in `docker-compose.test.yml`
- **Build Tags**: **DO NOT USE** - directory-based separation only

#### E2E Test Architecture
**E2E tests run the whole stack in Docker Compose but execute Go tests OUTSIDE the containers:**

- **Services**: Run in isolated Docker Compose environment (`docker-compose.test.yml`)
- **Test Runner**: Go tests execute on host machine, connecting to containerized services
- **Service URLs**: Tests connect to exposed ports (e.g., `http://localhost:18082`, `http://localhost:18081`)
- **Isolation**: Each test run gets fresh containers with clean state

#### E2E Test Structure
**ALL tests MUST use Given/When/Then structure for readability:**

```go
package e2e_test

import (
    "testing"
    "net/http"
)

func TestWorkflow_CompleteUserJourney(t *testing.T) {
    // given - assumes docker-compose.test.yml services are running
    client := &http.Client{}
    userData := buildUserData()
    
    // when - connect to containerized services
    response := makeRequest(client, "POST", "http://localhost:18081/api/register", userData)
    
    // then
    assert.Equal(t, http.StatusCreated, response.StatusCode)
    
    // Continue workflow testing with external service connections...
}
```

### 4. Browser-Based Frontend Tests (CRITICAL FOR HTMX)
- **Location**: `test/e2e/browser/` directory  
- **Package Name**: `package browser_test`
- **Dependencies**: chromedp for browser automation (MANDATORY for real frontend testing)
- **Scope**: Test complete frontend user interactions including HTMX workflows
- **Speed**: Slow (1s+ per test)
- **Browser**: Chrome/Chromium headless via Chrome DevTools Protocol
- **Build Tags**: **DO NOT USE** - directory-based separation only
- **CRITICAL**: These are the ONLY tests that validate real user experience

#### Browser Test Architecture
**Browser tests use chromedp to automate real browser interactions:**

- **Browser**: Chrome/Chromium headless instance via DevTools Protocol
- **Test Runner**: Go tests execute on host machine, controlling browser
- **Service URLs**: Tests connect to running services (e.g., `http://localhost:18082`)
- **Isolation**: Fresh browser context for each test

#### Browser Test Structure with chromedp
**ALL browser tests MUST use Given/When/Then structure:**

```go
package browser_test

import (
    "context"
    "testing"
    "github.com/chromedp/chromedp"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/assert"
)

func TestFrontend_HTMXDashboardLogin(t *testing.T) {
    // Skip browser tests in short mode
    if testing.Short() {
        t.Skip("Skipping browser tests in short mode")
    }
    
    // given - fresh browser context and test user
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    testEmail := "test@example.com"
    testPassword := "testpass123"
    var dashboardTitle string
    
    // when - perform complete login workflow
    err := chromedp.Run(ctx,
        // Navigate to login page
        chromedp.Navigate("http://localhost:18082/login"),
        chromedp.WaitVisible("#email", chromedp.ByID),
        
        // Fill login form
        chromedp.SendKeys("#email", testEmail),
        chromedp.SendKeys("#password", testPassword),
        chromedp.Click("#login-button"),
        
        // Wait for dashboard to load
        chromedp.WaitVisible("#dashboard", chromedp.ByID),
        chromedp.Text("h1", &dashboardTitle),
    )
    
    // then
    require.NoError(t, err)
    assert.Equal(t, "Dashboard", dashboardTitle)
}
```

#### HTMX Testing with chromedp
**Special patterns for testing HTMX applications:**

##### HTMX Request Lifecycle Testing
```go
func TestHTMX_SearchWithRequestLifecycle(t *testing.T) {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    var searchResults string
    
    err := chromedp.Run(ctx,
        chromedp.Navigate("http://localhost:18082/dashboard"),
        chromedp.WaitVisible("[data-testid='search-input']"),
        
        // Type in search box (triggers HTMX request)
        chromedp.SendKeys("[data-testid='search-input']", "AAPL"),
        
        // Wait for HTMX request to start (htmx-request class appears)
        chromedp.WaitVisible(".htmx-request"),
        
        // Wait for HTMX request to complete (htmx-request class disappears)
        WaitForHTMXComplete("[data-testid='search-widget']"),
        
        // Verify results
        chromedp.Text("[data-testid='search-results']", &searchResults),
    )
    
    require.NoError(t, err)
    assert.Contains(t, searchResults, "AAPL")
}

// Custom helper for HTMX completion
func WaitForHTMXComplete(selector string) chromedp.Action {
    return chromedp.Poll(`(selector) => {
        const el = document.querySelector(selector);
        return el && !el.classList.contains('htmx-request');
    }`, selector)
}
```

##### HTMX Event-Based Testing
```go
func TestHTMX_WatchlistManagement(t *testing.T) {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    err := chromedp.Run(ctx,
        // Login and navigate to dashboard
        loginToDashboard(ctx, "test@example.com", "testpass123"),
        
        // Search for symbol
        chromedp.SendKeys("[data-testid='search-input']", "AAPL"),
        WaitForHTMXComplete("[data-testid='search-widget']"),
        
        // Add to watchlist
        chromedp.Click("[data-testid='add-aapl-button']"),
        
        // Wait for watchlist to update via HTMX
        WaitForHTMXComplete("[data-testid='watchlist-widget']"),
        
        // Verify symbol appears in watchlist
        chromedp.WaitVisible("[data-testid='watchlist-aapl']"),
    )
    
    require.NoError(t, err)
}
```

#### HTMX Testing Requirements (MANDATORY)
**Critical considerations for HTMX applications - ONLY WAY TO TEST FRONTEND:**

1. **HTMX Request Lifecycle Testing** (REQUIRED):
   - **Wait for HTMX completion**: Always wait for `.htmx-request` class to disappear
   - **Monitor HTMX events**: Test htmx:beforeRequest, htmx:afterRequest, htmx:responseError
   - **Handle partial page updates**: Verify DOM changes from HTMX responses
   - **Test request indicators**: Verify loading states during HTMX requests

2. **HTMX Authentication Testing** (CRITICAL):
   - **Real browser cookies**: Test JWT cookies work in actual browser context
   - **Session persistence**: Verify authentication survives page interactions
   - **HTMX auth headers**: Ensure HTMX requests include proper authorization
   - **Auth failure handling**: Test 401/403 responses in HTMX context

3. **HTMX UI Interaction Testing** (USER EXPERIENCE):
   - **Use data-testid attributes**: Prefer semantic test IDs over CSS selectors
   - **Handle dynamic content**: Wait for elements to be visible before interaction
   - **Test auto-refresh**: Verify `hx-trigger="every Xs"` behavior
   - **Form submissions**: Test hx-post, hx-get form interactions
   - **Search-as-you-type**: Test hx-trigger="keyup" with debouncing

4. **HTMX Error Handling** (RELIABILITY):
   - **Network failures**: Test HTMX behavior when backend unavailable
   - **Timeout handling**: Test slow HTMX responses
   - **Error responses**: Test HTMX error responses and fallback behavior
   - **Graceful degradation**: Verify non-JS fallbacks work

#### Browser Testing Tool Selection (2025)
**chromedp was selected as the optimal browser automation framework for Go + HTMX applications:**

**Why chromedp:**
- **Native Go Integration**: Seamless integration with Go testing ecosystem
- **HTMX Compatibility**: Sufficient capabilities to handle HTMX async patterns
- **Direct DevTools Protocol**: No external WebDriver dependencies
- **Active Maintenance**: Strong Go community support and development
- **Performance**: Direct Chrome communication for fast test execution

**HTMX-Specific Testing Patterns:**
```go
// Wait for HTMX request to complete
func WaitForHTMXComplete(selector string) chromedp.Action {
    return chromedp.Poll(`(selector) => {
        const el = document.querySelector(selector);
        return el && !el.classList.contains('htmx-request');
    }`, selector)
}

// Monitor HTMX events
chromedp.WaitVisible(".htmx-request")  // Request started
WaitForHTMXComplete("#search-results") // Request completed
```

#### Browser Test Helpers
**Required helper functions for browser tests:**

```go
// Helper for login workflow
func loginToDashboard(ctx context.Context, email, password string) chromedp.Tasks {
    return chromedp.Tasks{
        chromedp.Navigate("http://localhost:18082/login"),
        chromedp.WaitVisible("#email"),
        chromedp.SendKeys("#email", email),
        chromedp.SendKeys("#password", password),
        chromedp.Click("#login-button"),
        chromedp.WaitVisible("#dashboard"),
    }
}

// Helper for waiting on HTMX events
func WaitForHTMXEvent(eventName string) chromedp.Action {
    return chromedp.Poll(`(eventName) => {
        return window.htmxEventReceived === eventName;
    }`, eventName)
}
```

## Test Organization

### File Naming
- Unit tests: `<component>_test.go` in same directory as code
- Integration tests: `test/integration/<component>_integration_test.go`
- E2E tests (HTTP): `test/e2e/<workflow>_e2e_test.go`
- Browser tests: `test/e2e/browser/<workflow>_browser_test.go`

### Clean Test Structure Patterns (ENHANCED)

#### Fluent Test Builder Pattern
**Based on TVN-cue-crab reference implementation - use fluent interfaces for readable test composition:**

```go
package e2e_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

type parts struct {
    *testing.T
    responseCode int
    responseBody []byte
    videoID      string
    checks       *checks
}

func newParts(t *testing.T) (*parts, *parts, *parts) {
    t.Helper()
    
    parts := &parts{
        T:      t,
        checks: newChecks(),
    }
    
    // Setup test environment
    parts.setupTestEnvironment()
    
    return parts, parts, parts  // given, when, then
}

// Fluent chaining methods
func (p *parts) and() *parts {
    return p
}

func (p *parts) uploadVideoFile(filename, title string) *parts {
    // Implementation
    return p
}

func (p *parts) videoUploadedSuccessfully() *parts {
    assert.Equal(p.T, http.StatusCreated, p.responseCode)
    return p
}

func (p *parts) noError() *parts {
    assert.NoError(p.T, p.checks.repeatRun())
    return p
}
```

#### Test Usage Example
**Clean, readable test structure with fluent chaining:**

```go
func TestVideoUploadAndRetrieval(t *testing.T) {
    videoFilePath := "./assets/test-video.mp4"
    given, when, then := newParts(t)

    when.
        uploadVideoFile(videoFilePath, "E2E Test Video")

    then.
        videoUploadedSuccessfully().and().
        getVideoByID(then.videoID).and().
        videoFoundSuccessfully()
}

func TestAssetEventProcessing(t *testing.T) {
    videoFilePath := "./assets/test-video.mp4"
    given, when, then := newParts(t)

    given.
        uploadVideoFile(videoFilePath, "Asset Event E2E Test Video").and().
        videoUploadedSuccessfully().and().
        getVideoByID(given.videoID)

    when.
        sendAssetEvent("ACTIVATE", given.videoID).and().
        messageSentSuccessfully()

    then.
        videoHasStatus("ready").and().
        videoHasAssets(2).and().
        noError()
}
```

#### Asynchronous Check Pattern
**Robust pattern for testing eventual consistency with retry logic:**

```go
type checks struct {
    interval time.Duration
    timeout  time.Duration
    checks   []*check
}

type check struct {
    f   func() error
    ok  bool
    err error
}

func newChecks() *checks {
    // CI-aware defaults
    interval := 500 * time.Millisecond
    timeout := 20 * time.Second
    
    if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
        interval = 2 * time.Second
        timeout = 60 * time.Second
    }
    
    return &checks{
        interval: interval,
        timeout:  timeout,
        checks:   make([]*check, 0, 10),
    }
}

func (cs *checks) add(f func() error) {
    cs.checks = append(cs.checks, &check{f: f})
}

func (cs *checks) repeatRun() error {
    tick := time.Tick(cs.interval)
    limit := time.After(cs.timeout)
    
    for {
        select {
        case <-limit:
            return cs.error()
        case <-tick:
            cs.run()
            if cs.ok() {
                return nil
            }
        }
    }
}

// Usage in test methods
func (p *parts) videoHasStatus(expectedStatus string) *parts {
    p.checks.add(func() error {
        resp, err := p.httpClient.Get("http://" + p.envs.ServiceAddress + "/videos/" + p.videoID)
        if err != nil {
            return fmt.Errorf("get video by ID request: %w", err)
        }
        defer resp.Body.Close()
        
        // Parse and validate status
        var videoResp struct { Status string `json:"status"` }
        if err := json.NewDecoder(resp.Body).Decode(&videoResp); err != nil {
            return fmt.Errorf("parse video response: %w", err)
        }
        
        if videoResp.Status != expectedStatus {
            return fmt.Errorf("expected status %s, got %s", expectedStatus, videoResp.Status)
        }
        
        return nil
    })
    return p
}
```

#### Template-Based Test Data Pattern
**Clean approach for managing complex test data with templates:**

```go
// Template files in templates/ directory
// templates/asset_event.json
{
    "Type": "{{.Type}}",
    "TenantUUID": "{{.TenantUUID}}",
    "ExternalUUID": "{{.ExternalUUID}}",
    "Name": "{{.Name}}",
    "Active": {{.Active}},
    "Since": "{{.Since}}",
    "TechnicalTags": "{{.TechnicalTags}}"
}

// Test method using templates
func (p *parts) sendAssetEvent(eventType, videoID string) *parts {
    eventTime := time.Now().Add(1 * time.Second).Format(time.RFC3339)
    return p.sendMessageFromTemplate("asset_event.json", map[string]any{
        "Type":          eventType,
        "TenantUUID":    "test-tenant-uuid",
        "ExternalUUID":  videoID,
        "Name":          "E2E Test Video",
        "Active":        true,
        "Since":         eventTime,
        "TechnicalTags": "",
    })
}

func (p *parts) sendMessageFromTemplate(templateFile string, data map[string]any) *parts {
    tmpl, err := template.ParseFiles(filepath.Join("templates", templateFile))
    assert.NoError(p.T, err, "parse template file")

    var buf bytes.Buffer
    err = tmpl.Execute(&buf, data)
    assert.NoError(p.T, err, "execute template")

    return p.sendSQSMessage(buf.String())
}
```

### Test Helper Functions
**ONLY helper functions MUST use `t.Helper()` (NOT test functions).** All helpers should follow Given/When/Then structure:

```go
func setupTestData(t *testing.T) TestData {
    t.Helper()
    
    // given
    data := TestData{
        Field1: "value1",
        Field2: "value2",
    }
    
    return data
}

// Enhanced fluent helper
func (p *parts) cleanupDatabase() *parts {
    if p.db == nil {
        p.Log("Warning: database connection is nil, skipping cleanup")
        return p
    }

    queries := []string{
        "DELETE FROM crab.assets",
        "DELETE FROM crab.transcodings", 
        "DELETE FROM crab.videos",
    }

    for _, query := range queries {
        _, err := p.db.Exec(query)
        if err != nil {
            p.Logf("Warning: failed to cleanup table with query '%s': %v", query, err)
        }
    }
    
    return p
}
```

## Running Tests

### Commands
```bash
# Unit tests only (fast feedback)
make test-unit

# Integration tests with testcontainers
make test-integration  

# E2E tests (HTTP-based) with docker-compose.test.yml
make test-e2e

# Browser tests with chromedp (CRITICAL for frontend validation)
make test-browser

# All tests (unit + integration + e2e + browser) 
make test-all

# Pre-commit check (unit tests + lint only - fast)
make check

# IMPORTANT: Browser tests are MANDATORY for frontend features
# HTTP-based E2E tests DO NOT validate frontend functionality
```

### Test Execution
- Unit tests: `go test ./internal/... ./pkg/...` (excludes `test/` directories)
- Integration tests: `go test ./test/integration/...`
- E2E tests (HTTP): 
  - `make test-e2e-up` - Start docker-compose.test.yml services
  - `make test-e2e-run` - Run `go test ./test/e2e/...` against running services
  - `make test-e2e-down` - Stop docker-compose.test.yml services
  - `make test-e2e` - Complete workflow (up → run → down)
- Browser tests: 
  - Requires running services (`make test-e2e-up` or `make up`)
  - `go test ./test/e2e/browser/...` - Run browser tests with chromedp
  - Browser tests download Chrome automatically if not present

**IMPORTANT: Do NOT use build tags for tests!** Test separation is achieved through directory structure only.

## Mocking Guidelines

### Mock Creation
Use `testify/mock` for all mocks:
```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetUser(id string) (*User, error) {
    args := m.Called(id)
    if user := args.Get(0); user != nil {
        return user.(*User), args.Error(1)
    }
    return nil, args.Error(1)
}
```

### Mock Assertions
Always verify mock expectations:
```go
mockRepo.On("GetUser", "123").Return(expectedUser, nil)
// ... test code ...
mockRepo.AssertExpectations(t)
```

## Test Data Management

### Unit Tests
- Use in-memory data structures
- Create minimal test fixtures
- No external file dependencies

### Integration Tests  
- Use testcontainers for clean state
- Seed databases with test data
- Clean up after each test

### E2E Tests
- Use docker-compose.test.yml volumes
- Reset state between test runs
- Test with realistic data sets

### Test Environment Management (ENHANCED)

#### CI-Aware Test Configuration
**Automatically adapt test behavior based on execution environment:**

```go
func newChecks() *checks {
    // Default values for local development
    interval := 500 * time.Millisecond
    timeout := 20 * time.Second

    // Use longer intervals for CI environments
    if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
        interval = 2 * time.Second
        timeout = 60 * time.Second
    }

    return &checks{
        interval: interval,
        timeout:  timeout,
        checks:   make([]*check, 0, 10),
    }
}
```

#### Environment-Specific Test Behavior
**Gracefully handle different execution environments:**

```go
func (p *parts) cleanupDatabase() {
    if p.db == nil {
        p.Log("Warning: database connection is nil, skipping cleanup")
        return
    }

    queries := []string{
        "DELETE FROM crab.assets",
        "DELETE FROM crab.transcodings",
        "DELETE FROM crab.videos",
    }

    for _, query := range queries {
        _, err := p.db.Exec(query)
        if err != nil {
            p.Logf("Warning: failed to cleanup table with query '%s': %v", query, err)
        }
    }
}

func (p *parts) cleanupSQSQueue() {
    if p.sqsClient == nil {
        p.Log("Warning: SQS client is nil, skipping SQS cleanup")
        return
    }

    queueUrl, err := url.JoinPath(p.sqsConfig.QueueUrl, p.sqsConfig.QueueName)
    if err != nil {
        p.Logf("Warning: failed to build queue URL for cleanup: %v", err)
        return
    }

    _, err = p.sqsClient.PurgeQueue(p.ctx, &awssqs.PurgeQueueInput{
        QueueUrl: aws.String(queueUrl),
    })
    if err != nil {
        p.Logf("Warning: failed to purge SQS queue: %v", err)
    }
}
```

#### Test Asset Management
**Organize test files and assets cleanly:**

```
test/
├── e2e/
│   ├── assets/           # Test data files (videos, images, etc.)
│   │   └── test-video.mp4
│   ├── templates/        # Template files for test data generation
│   │   ├── asset_event.json
│   │   └── encode_task_event.json
│   ├── .env.test        # Test environment configuration
│   ├── docker-compose.test.yml
│   ├── e2e_test.go      # Main test scenarios
│   ├── e2e_parts_test.go # Test helpers and fluent interface
│   └── e2e_checks_test.go # Asynchronous check implementation
└── integration/
    └── ...
```

#### Configuration Management
**Clean separation of test configurations:**

```go
// .env.test
TEST_DB_URL=postgres://test:test@localhost:5433/testdb
TEST_SQS_QUEUE_URL=http://localhost:4566/000000000000/test-queue
TEST_SERVICE_ADDRESS=http://localhost:18082

// Test setup with environment loading
func newParts(t *testing.T) (*parts, *parts, *parts) {
    t.Helper()
    var envs config.Environment
    
    // Load test-specific environment
    err := godotenv.Load(".env.test")
    assert.NoError(t, err, "load test config")

    err = env.Parse(&envs)
    assert.NoError(t, err, "parse test config")
    
    // Continue with test setup...
}
```

## Test Readability and Maintainability (ENHANCED)

### Clean Test Structure Principles
**Based on TVN-cue-crab reference implementation - prioritize readability and maintainability:**

#### 1. Fluent Interface Pattern
**Use method chaining for readable test workflows:**

```go
// Clean, readable test with fluent chaining
func TestVideoUploadAndRetrieval(t *testing.T) {
    videoFilePath := "./assets/test-video.mp4"
    given, when, then := newParts(t)

    when.
        uploadVideoFile(videoFilePath, "E2E Test Video")

    then.
        videoUploadedSuccessfully().and().
        getVideoByID(then.videoID).and().
        videoFoundSuccessfully()
}

// Complex workflow with multiple steps
func TestAssetEventProcessing(t *testing.T) {
    videoFilePath := "./assets/test-video.mp4"
    given, when, then := newParts(t)

    given.
        uploadVideoFile(videoFilePath, "Asset Event E2E Test Video").and().
        videoUploadedSuccessfully().and().
        getVideoByID(given.videoID)

    when.
        sendAssetEvent("ACTIVATE", given.videoID).and().
        messageSentSuccessfully()

    then.
        videoHasStatus("ready").and().
        videoHasAssets(2).and().
        noError()
}
```

#### 2. Given/When/Then Structure Enforcement
**ALL tests must follow the Given/When/Then pattern:**

```go
func TestEndpointHealth(t *testing.T) {
    _, when, then := newParts(t)

    when.
        checkHealth()

    then.
        responseCodeIs(200)
}

// Tests with setup (given) use all three parts
func TestComplexWorkflow(t *testing.T) {
    given, when, then := newParts(t)

    given.
        setupTestData().and().
        configureServices()

    when.
        executeWorkflow()

    then.
        verifyExpectedResult().and().
        noError()
}
```

#### 3. Self-Documenting Test Methods
**Method names should clearly describe their purpose:**

```go
// ✅ Clear, descriptive method names
func (p *parts) videoUploadedSuccessfully() *parts
func (p *parts) videoHasStatus(expectedStatus string) *parts
func (p *parts) sendAssetEvent(eventType, videoID string) *parts
func (p *parts) messageSentSuccessfully() *parts

// ❌ Vague or unclear method names
func (p *parts) checkVideo() *parts
func (p *parts) doEvent() *parts
func (p *parts) verify() *parts
```

#### 4. Graceful Error Handling
**Provide meaningful error messages and handle failures gracefully:**

```go
func (p *parts) cleanupDatabase() {
    if p.db == nil {
        p.Log("Warning: database connection is nil, skipping cleanup")
        return
    }

    queries := []string{
        "DELETE FROM crab.assets",
        "DELETE FROM crab.transcodings",
        "DELETE FROM crab.videos",
    }

    for _, query := range queries {
        _, err := p.db.Exec(query)
        if err != nil {
            p.Logf("Warning: failed to cleanup table with query '%s': %v", query, err)
        }
    }
}

func (p *parts) videoHasStatus(expectedStatus string) *parts {
    p.checks.add(func() error {
        resp, err := p.httpClient.Get("http://" + p.envs.ServiceAddress + "/videos/" + p.videoID)
        if err != nil {
            return fmt.Errorf("get video by ID request: %w", err)
        }
        defer resp.Body.Close()

        // ... validation logic with detailed error messages
        if videoResp.Status != expectedStatus {
            return fmt.Errorf("expected status %s, got %s", expectedStatus, videoResp.Status)
        }

        return nil
    })
    return p
}
```

#### 5. Test Data Management
**Organize test data cleanly with templates and fixtures:**

```go
// Template-based test data generation
func (p *parts) sendAssetEvent(eventType, videoID string) *parts {
    eventTime := time.Now().Add(1 * time.Second).Format(time.RFC3339)
    return p.sendMessageFromTemplate("asset_event.json", map[string]any{
        "Type":          eventType,
        "TenantUUID":    "test-tenant-uuid",
        "ExternalUUID":  videoID,
        "Name":          "E2E Test Video",
        "Active":        true,
        "Since":         eventTime,
        "TechnicalTags": "",
    })
}

// Template file: templates/asset_event.json
{
    "Type": "{{.Type}}",
    "TenantUUID": "{{.TenantUUID}}",
    "ExternalUUID": "{{.ExternalUUID}}",
    "Name": "{{.Name}}",
    "Active": {{.Active}},
    "Since": "{{.Since}}",
    "TechnicalTags": "{{.TechnicalTags}}"
}
```

#### 6. Asynchronous Testing Patterns
**Handle eventual consistency with robust retry logic:**

```go
// CI-aware retry configuration
func newChecks() *checks {
    interval := 500 * time.Millisecond
    timeout := 20 * time.Second

    // Use longer intervals for CI environments
    if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
        interval = 2 * time.Second
        timeout = 60 * time.Second
    }

    return &checks{
        interval: interval,
        timeout:  timeout,
        checks:   make([]*check, 0, 10),
    }
}

// Accumulate checks and execute with retry
func (p *parts) videoHasAssets(expectedCount int) *parts {
    p.checks.add(func() error {
        query := `SELECT COUNT(*) FROM crab.assets WHERE video_id = $1`
        
        // Database query with retry logic
        var assetCount int
        err := p.db.Get(&assetCount, query, videoUUID)
        if err != nil {
            return fmt.Errorf("failed to query asset count: %w", err)
        }

        if assetCount != expectedCount {
            return fmt.Errorf("expected %d assets, got %d", expectedCount, assetCount)
        }

        return nil
    })
    return p
}

// Execute all accumulated checks with retry
func (p *parts) noError() *parts {
    assert.NoError(p.T, p.checks.repeatRun(), "checks failed")
    return p
}
```

#### 7. Test Organization Best Practices
**Structure test files for maximum maintainability:**

```
test/
├── e2e/
│   ├── assets/               # Test data files
│   │   └── test-video.mp4
│   ├── templates/            # Template files for test data
│   │   ├── asset_event.json
│   │   └── encode_task_event.json
│   ├── .env.test           # Test environment configuration
│   ├── docker-compose.test.yml
│   ├── e2e_test.go         # Main test scenarios (readable workflows)
│   ├── e2e_parts_test.go   # Test helpers and fluent interface
│   └── e2e_checks_test.go  # Asynchronous check implementation
```

#### 8. Environment Configuration Management
**Separate test configurations cleanly:**

```go
// .env.test
TEST_DB_URL=postgres://test:test@localhost:5433/testdb
TEST_SQS_QUEUE_URL=http://localhost:4566/000000000000/test-queue
TEST_SERVICE_ADDRESS=http://localhost:18082

// Test setup with environment loading
func newParts(t *testing.T) (*parts, *parts, *parts) {
    t.Helper()
    var envs config.Environment
    
    // Load test-specific environment
    err := godotenv.Load(".env.test")
    assert.NoError(t, err, "load test config")

    err = env.Parse(&envs)
    assert.NoError(t, err, "parse test config")
    
    // Continue with test setup...
}
```

### Test Anti-Patterns to Avoid

#### ❌ Anti-Pattern: Complex Test Setup
```go
// Bad: Test setup mixed with test logic
func TestBadExample(t *testing.T) {
    // Complex setup mixed in
    db, err := sqlx.Connect("postgres", "postgres://test:test@localhost:5433/testdb")
    if err != nil {
        t.Fatal(err)
    }
    
    httpClient := &http.Client{}
    sqsClient := setupSQSClient()
    
    // Test logic buried in setup
    resp, err := httpClient.Post("http://localhost:8080/videos", "application/json", bytes.NewReader(data))
    if err != nil {
        t.Fatal(err)
    }
    
    // Assertions mixed with setup
    if resp.StatusCode != 201 {
        t.Errorf("expected 201, got %d", resp.StatusCode)
    }
}
```

#### ✅ Correct Pattern: Clean Separation
```go
// Good: Clean separation with fluent interface
func TestGoodExample(t *testing.T) {
    videoFilePath := "./assets/test-video.mp4"
    _, when, then := newParts(t)

    when.
        uploadVideoFile(videoFilePath, "Test Video")

    then.
        videoUploadedSuccessfully()
}
```

## Quality Standards

### Coverage
- Unit tests: Aim for >90% coverage of public APIs
- Integration tests: Cover critical integration points  
- E2E tests: Cover main user workflows

### Performance
- Unit tests: <10ms each
- Integration tests: <1s each
- E2E tests: <30s each

### Reliability
- Tests must be deterministic (no flaky tests)
- No dependencies on external internet services
- Use fixed test data, not random generation
- **ALWAYS prioritize fixing failing tests over skipping them**
- **ZERO TOLERANCE**: Failing, skipped, flaky, or timing out tests are unacceptable
- **100% RELIABILITY STANDARD**: All tests must pass consistently across environments
- **NO CONDITIONAL TESTS**: Never skip tests based on runtime conditions
- **IMMEDIATE FAILURE ACTION**: Any test failure must be fixed before proceeding with development

## TDD Process

### Red-Green-Refactor
1. **Red**: Write failing test first
2. **Green**: Write minimal code to pass test
3. **Refactor**: Improve code while keeping tests green
4. **Repeat**: Continue cycle for all features

### Test-First Development
- Write unit test before implementation
- Write integration test for component interactions
- Write E2E test for complete workflows
- Never skip tests - they are part of the implementation

## Continuous Integration

### Pre-commit Requirements
**`make check` is the main development validation command and git hook validation.**

`make check` runs quickly (< 30 seconds) with essential validations:
```bash
make check  # Runs: vet, test-unit, lint
```

**Components of `make check`:**
- **`go vet`** - Static analysis for common errors
- **`make test-unit`** - Fast unit tests only (no external dependencies)
- **`make lint`** - golangci-lint comprehensive code quality checks

**NOT included in `make check` (run separately):**
- `make test-integration` - Slower integration tests with testcontainers
- `make test-e2e` - Full system tests with Docker Compose

### Test Execution Order
1. Unit tests (fastest)
2. Integration tests (moderate)
3. E2E tests (slowest)

Fail fast - stop on first test failure in CI.

## ZERO TOLERANCE POLICY

### Unacceptable Test States
**The following test states are strictly prohibited and must be fixed immediately:**

- **❌ FAILING TESTS**: Any test that returns non-zero exit code
- **❌ SKIPPED TESTS**: Any test marked with `t.Skip()` or conditional execution
- **❌ FLAKY TESTS**: Tests that pass/fail inconsistently across runs
- **❌ TIMEOUT TESTS**: Tests that exceed timeout limits or hang indefinitely

### Enforcement Rules
- **BLOCKING**: No commits allowed with any of the above test states
- **IMMEDIATE FIX**: All failing tests must be resolved before continuing development
- **NO WORKAROUNDS**: Skipping or ignoring tests is never acceptable
- **100% PASS RATE**: Only tests that consistently pass in all environments are acceptable

### Developer Responsibilities
1. **Test Before Commit**: Always run `make check` and `make test-all` before committing
2. **Fix Don't Skip**: Address test failures through code fixes, never by disabling tests
3. **Report Issues**: Immediately escalate infrastructure issues causing test instability
4. **Zero Tolerance Mindset**: Maintain highest standards for test reliability

This policy ensures production-ready code quality and reliable automated testing pipelines.
