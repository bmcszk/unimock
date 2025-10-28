# Fluent Test Design Standards (MANDATORY)

This document provides comprehensive standards for writing fluent E2E tests using the Given/When/Then pattern. All E2E tests MUST follow these exact patterns and rules.

## Fluent Test Builder Pattern
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

### Test Usage Example
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

## E2E Test Structure Requirements

**ALL E2E tests MUST follow the exact Given/When/Then structure with these specific patterns:**

```go
func TestWorkflow_CriticalUserAction(t *testing.T) {
    given, when, then := newParts(t)

    given.
        setupTestData().and().
        configureServices()

    when.
        executeCriticalAction()

    then.
        actionSucceeded().and().
        noError()
}
```

## MANDATORY Formatting Rules

### 1. Given/When/Then Section Structure
**Every E2E test MUST use this exact pattern:**

```go
// ✅ CORRECT: Proper Given/When/Then structure
func TestAPI_ResourceCreation(t *testing.T) {
    given, when, then := newParts(t)

    given.
        setupResourceScenario()

    when.
        createResourceViaAPI()

    then.
        resourceCreatedSuccessfully().and().
        responseContainsExpectedData()
}

// ❌ FORBIDDEN: Missing then section
func TestAPI_ResourceCreation(t *testing.T) {
    given, when, _ := newParts(t)

    given.
        setupResourceScenario()

    when.
        createResourceViaAPI().and().
        resourceCreatedSuccessfully()  // WRONG: validation in when section
}
```

### 2. Fluent Method Chaining Format
**ALL fluent chains must follow this exact formatting:**

```go
// ✅ CORRECT: Proper fluent formatting
given.
    setupScenario().and().
    configureServices()

when.
    executeAction()

then.
    verifyResult().and().
    cleanupState()
```

### 3. Method Positioning Rules
**Follow these rules for method chaining:**

```go
// ✅ CORRECT: .and() at end of line, continuation on next line
when.
    executeHTTPRequest("/api/resource").and().
    validateResponseCode(201)

// ✅ CORRECT: Multiple validations chained
then.
    responseCodeIs(200).and().
    responseContains("success").and().
    noError()

// ❌ FORBIDDEN: .and() at start of line
when.
    executeHTTPRequest("/api/resource")
        .and().validateResponseCode(201)  // WRONG: .and() mispositioned
```

## Variable Declaration Rules

### 1. Section Variable Declaration
**Declare variables based on what sections are actually used:**

```go
// ✅ CORRECT: Only declare variables that are used
func TestWithGivenAndWhen(t *testing.T) {
    given, when, then := newParts(t)
    // Uses all three sections
}

func TestWithWhenAndThen(t *testing.T) {
    _, when, then := newParts(t)
    // No setup needed
}

func TestWithWhenOnly(t *testing.T) {
    _, when, _ := newParts(t)
    // Simple test with no verification
}
```

### 2. Unused Variable Elimination
**NEVER declare unused variables:**

```go
// ❌ FORBIDDEN: Unused given variable
func TestSimpleAction(t *testing.T) {
    given, when, then := newParts(t)  // given is unused
    when.
        executeAction()
    then.
        verifyResult()
}

// ✅ CORRECT: Only declare what's needed
func TestSimpleAction(t *testing.T) {
    _, when, then := newParts(t)
    when.
        executeAction()
    then.
        verifyResult()
}
```

## Response Handling Patterns

### 1. Response Storage Pattern
**Store responses in the parts struct for better readability:**

```go
// ✅ CORRECT: Store responses in parts struct
func (p *parts) executeHTTPRequest(endpoint string) *parts {
    responses, err := p.httpClient.Get(endpoint)
    p.require.NoError(err, "Failed to execute HTTP request")
    p.responses = make([]any, len(responses))
    for i, resp := range responses {
        p.responses[i] = resp
    }
    return p
}

func (p *parts) validateResponse() *parts {
    // Access stored responses from parts struct
    resp := p.responses[0].(*http.Response)
    p.require.Equal(200, resp.StatusCode)
    return p
}
```

### 2. Method Return Value Pattern
**Fluent methods should return *parts for chaining:**

```go
// ✅ CORRECT: Methods support fluent chaining
func (p *parts) createResource() *parts {
    // Implementation
    return p
}

// ✅ CORRECT: Usage in fluent chain
when.
    createResource().and().
    verifyResource()
```

## Migration Guidelines: Standard to E2E Tests

### 1. Migration Decision Matrix
**Use this guide to determine when to create E2E tests:**

| Test Need | Recommended Approach | Migration Strategy |
|-----------|---------------------|-----------------|
| Single function unit logic | **Unit Test** | Add unit tests if missing |
| Service interaction | **Integration Test** | Create integration tests |
| Complete user workflow | **E2E Test** | Create E2E test with Given/When/Then |
| API contract validation | **E2E Test** | Convert integration tests to E2E |
| Frontend user journey | **Browser Test** | Create browser tests with chromedp |

### 2. Standard Test to E2E Migration Process

**Step 1: Identify Migration Candidates**
```bash
# Find integration tests that cover user workflows
find test/integration -name "*integration_test.go" | \
    xargs grep -l "workflow\|journey\|end-to-end"
```

**Step 2: Analyze Current Test Structure**
```go
// BEFORE: Standard integration test
func TestUserService_CreateUser(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()

    service := NewUserService(db)
    user := &User{Name: "Test User", Email: "test@example.com"}

    created, err := service.CreateUser(user)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, created.Name)
    assert.NotZero(t, created.ID)
}
```

**Step 3: Convert to Given/When/Then Structure**
```go
// AFTER: Migrated to E2E with fluent pattern
func TestUserService_CreateUser(t *testing.T) {
    given, when, then := newParts(t)

    given.
        setupTestDatabase().and().
        initializeUserService()

    when.
        createUserWithName("Test User").and().
        createUserWithEmail("test@example.com")

    then.
        userCreatedSuccessfully().and().
        userHasCorrectName("Test User").and().
        userHasValidID()
}
```

### 3. E2E Test Implementation Template
**Use this template for new E2E tests:**

```go
package e2e_test

import (
    "testing"
    "github.com/stretchr/testify/require"
)

type parts struct {
    *testing.T
    require     *require.Assertions
    // Add test-specific fields
    userID      string
    response    *http.Response
    httpClient  *http.Client
}

func newParts(t *testing.T) (*parts, *parts, *parts) {
    t.Helper()

    parts := &parts{
        T:        t,
        require:  require.New(t),
        // Initialize test-specific fields
    }

    parts.setupTestEnvironment()
    return parts, parts, parts
}

// Fluent helper methods
func (p *parts) and() *parts {
    return p
}

func (p *parts) setupTestEnvironment() *parts {
    // Implementation
    return p
}

// Given section methods
func (p *parts) initializeUserService() *parts {
    // Implementation
    return p
}

// When section methods
func (p *parts) createUserWithDetails(details map[string]string) *parts {
    // Implementation
    return p
}

// Then section methods
func (p *parts) userCreatedSuccessfully() *parts {
    // Implementation
    return p
}

// Test implementation
func TestUserService_CompleteUserCreation(t *testing.T) {
    given, when, then := newParts(t)

    given.
        initializeUserService()

    when.
        createUserWithDetails(map[string]string{
            "name":  "John Doe",
            "email": "john@example.com",
        })

    then.
        userCreatedSuccessfully().and().
        userExistsInDatabase().and().
        userHasCorrectDetails()
}
```

## E2E Test Quality Standards

### 1. Test Completeness Requirements
**Every E2E test must cover:**
- ✅ Setup phase (given)
- ✅ Action phase (when)
- ✅ Verification phase (then)
- ✅ Cleanup (if needed)
- ✅ Error handling validation

### 2. Readability Standards
**Tests must be self-documenting:**
- ✅ Clear test names: `TestComponent_Action_ExpectedResult`
- ✅ Descriptive method names
- ✅ Proper section comments
- ✅ Meaningful assertion messages

### 3. Maintainability Requirements
**Tests must be maintainable:**
- ✅ Reusable helper methods
- ✅ Clear separation of concerns
- ✅ Consistent error handling
- ✅ Environment-independent execution

## Anti-Patterns to Avoid

### ❌ Structural Anti-Patterns
```go
// FORBIDDEN: Missing then section
func TestBad(t *testing.T) {
    given, when, _ := newParts(t)
    when.executeAction()  // No verification
}

// FORBIDDEN: Validation in wrong section
func TestBad(t *testing.T) {
    given, when, then := newParts(t)
    when.executeAction().and().verifyResult()  // Wrong section
}

// FORBIDDEN: Unused variables
func TestBad(t *testing.T) {
    given, when, then := newParts(t)  // given, then unused
    when.executeAction()
    then.verifyResult()
}
```

### ❌ Formatting Anti-Patterns
```go
// FORBIDDEN: Inconsistent chaining
when.
    method1().
        method2()  // Missing .and().
    then.verify()

// FORBIDDEN: Misplaced .and()
when.
    method1().
        .and().method2()  // Wrong indentation

// FORBIDDEN: Long single-line chains
when.method1().and().method2().and().method3().and().method4()  // Too long
```

## Advanced Patterns

### Asynchronous Check Pattern
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

### Template-Based Test Data Pattern
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

## Test Helper Functions
**ONLY helper functions MUST use `t.Helper()` (NOT test functions).** All helpers should follow Given/When/Then structure:

```go
func setupTestData(t *testing.T) TestData {
    t.Helper()

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

---

## Test Environment Management (ENHANCED)

### CI-Aware Test Configuration
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

### Environment-Specific Test Behavior
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

### Test Asset Management
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

### Configuration Management
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

        var count int
        err := p.db.QueryRow(query, p.videoID).Scan(&count)
        if err != nil {
            return fmt.Errorf("query assets count: %w", err)
        }

        if count != expectedCount {
            return fmt.Errorf("expected %d assets, got %d", expectedCount, count)
        }

        return nil
    })
    return p
}
```

#### 7. Test Organization Best Practices
**Structure tests for maximum maintainability:**

```
test/
├── e2e/
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
// Separate test configs with environment-specific behavior
func setupTestEnvironment(t *testing.T) *TestEnvironment {
    t.Helper()

    // Load test-specific config
    err := godotenv.Load(".env.test")
    require.NoError(t, err, "load test environment")

    env := &TestEnvironment{
        DBURL:     os.Getenv("TEST_DB_URL"),
        SQSQueue:  os.Getenv("TEST_SQS_QUEUE_URL"),
        ServiceAddr: os.Getenv("TEST_SERVICE_ADDRESS"),
    }

    return env
}
```

## Test Anti-Patterns to Avoid

### ❌ Anti-Pattern: Complex Test Setup
```go
// BAD: Complex setup mixed with test logic
func TestVideoProcessing(t *testing.T) {
    // 50 lines of complex setup here
    db := setupDatabase(t, "complex-config.sql")
    sqsClient := setupSQS(t)
    s3Client := setupS3(t)
    videoProcessor := NewVideoProcessor(db, sqsClient, s3Client)

    // Test logic buried in setup
    // ... actual test logic
}
```

### ✅ Correct Pattern: Clean Separation
```go
// GOOD: Clean setup separation
func TestVideoProcessing(t *testing.T) {
    given, when, then := newParts(t)

    given.
        setupVideoProcessingEnvironment().and().
        uploadTestVideo()

    when.
        startVideoProcessing()

    then.
        videoProcessedSuccessfully().and().
        assetsCreated()
}

func (p *parts) setupVideoProcessingEnvironment() *parts {
    // All complex setup encapsulated in helper method
    p.db = setupDatabase(p.T, "test-config.sql")
    p.sqsClient = setupSQS(p.T)
    p.s3Client = setupS3(p.T)
    p.videoProcessor = NewVideoProcessor(p.db, p.sqsClient, p.s3Client)
    return p
}
```

## Testing Commands Reference

### E2E Test Execution Commands
```bash
# E2E tests (HTTP-based) with docker-compose.test.yml
make test-e2e-up      # Start docker-compose.test.yml services
make test-e2e-run     # Run `go test ./test/e2e/...` against running services
make test-e2e-down    # Stop docker-compose.test.yml services
make test-e2e         # Complete workflow (up → run → down)

# All tests (unit + integration + e2e + browser)
make test-all

# Pre-commit check (unit tests + lint only - fast)
make check

# IMPORTANT: Browser tests are MANDATORY for frontend features
# HTTP-based E2E tests DO NOT validate frontend functionality
```

### Test Environment Variables
- `TEST_DB_URL` - Database connection string for tests
- `TEST_SQS_QUEUE_URL` - SQS queue URL for async tests
- `TEST_SERVICE_ADDRESS` - Service base URL for HTTP tests
- `CI` - Automatically set in CI environments
- `GITHUB_ACTIONS` - Automatically set in GitHub Actions

---

**These standards are MANDATORY for all E2E tests. No exceptions.**