# PRD: Universal HTTP Client Library Expansion

**Document ID:** REQ-CLIENT-UNIVERSAL-001  
**Version:** 1.0  
**Status:** Draft  
**Created:** 2025-06-24  
**Last Updated:** 2025-06-24  

## Executive Summary

Expand the existing Unimock client library to provide a universal HTTP client capable of making requests to any mock endpoint (not just scenario management endpoints). The client should provide HTTP method-specific functions with context support and return structured response objects.

## Background

The current Unimock client library (`pkg/client/`) only supports scenario management operations (Create, Get, List, Update, Delete scenarios). Users need a universal client that can interact with any mock endpoint to test their applications against mocked services.

### Current State
- Client exists at `pkg/client/client.go`
- Supports only scenario management API (`/_uni/scenarios`)
- Has context support and proper error handling
- Well-tested with comprehensive unit tests

### Problem Statement
Users want to use Unimock client to make HTTP requests to any mock endpoint, not just scenario management. This would enable:
1. Testing applications against mocked services programmatically
2. Setting up test data and then verifying responses
3. Using Unimock as both a test server and client in integration tests

## Requirements

### Functional Requirements

#### FR-1: HTTP Method Functions
- **Requirement:** Provide functions for all standard HTTP methods
- **Details:** 
  - `Get(ctx, path, headers) (*Response, error)`
  - `Head(ctx, path, headers) (*Response, error)` 
  - `Post(ctx, path, headers, body) (*Response, error)`
  - `Put(ctx, path, headers, body) (*Response, error)`
  - `Delete(ctx, path, headers) (*Response, error)`
  - `Patch(ctx, path, headers, body) (*Response, error)`
  - `Options(ctx, path, headers) (*Response, error)`

#### FR-2: Response Structure
- **Requirement:** Return structured response containing status code, headers, and body
- **Details:**
  ```go
  type Response struct {
      StatusCode int
      Headers    http.Header  
      Body       []byte
  }
  ```

#### FR-3: Context Support
- **Requirement:** All HTTP method functions must accept context.Context as first parameter
- **Details:** Enable request cancellation, timeouts, and deadline handling

#### FR-4: Flexible Parameters
- **Requirement:** Support optional parameters appropriate for each method
- **Details:**
  - GET, HEAD, DELETE, OPTIONS: context, path, headers (body not needed)
  - POST, PUT, PATCH: context, path, headers, body
  - Headers should be `map[string]string` or `http.Header`

#### FR-5: Backward Compatibility  
- **Requirement:** Maintain existing scenario management functions unchanged
- **Details:** Existing `CreateScenario`, `GetScenario`, etc. must continue working

### Technical Requirements

#### TR-1: Package Structure
- **Requirement:** Add new HTTP method functions to existing `client` package
- **Details:** Extend `pkg/client/client.go` without breaking changes

#### TR-2: Error Handling
- **Requirement:** Consistent error handling following existing patterns
- **Details:** 
  - Network errors return wrapped errors
  - HTTP error status codes (4xx, 5xx) return errors with status and body
  - Context cancellation properly handled

#### TR-3: HTTP Client Reuse
- **Requirement:** Reuse existing HTTP client configuration and patterns
- **Details:** Use same timeout, base URL handling, and client setup

#### TR-4: URL Construction
- **Requirement:** Handle relative and absolute paths correctly
- **Details:**
  - Relative paths (e.g., "/api/users") resolve against base URL
  - Absolute URLs override base URL
  - Proper URL encoding and path joining

### Quality Requirements

#### QR-1: Test Coverage
- **Requirement:** Comprehensive unit tests for all new functions
- **Details:** 
  - Test each HTTP method function
  - Test error conditions (network errors, HTTP errors)
  - Test context cancellation
  - Test URL construction edge cases

#### QR-2: Documentation
- **Requirement:** Clear godoc documentation for all public functions
- **Details:** Include usage examples and parameter descriptions

#### QR-3: Performance
- **Requirement:** No significant performance impact over raw http.Client
- **Details:** Minimal overhead from abstraction layer

## Design

### API Design

```go
// Response represents an HTTP response from the server
type Response struct {
    StatusCode int         // HTTP status code (200, 404, etc.)
    Headers    http.Header // Response headers
    Body       []byte      // Response body content
}

// HTTP method functions
func (c *Client) Get(ctx context.Context, path string, headers map[string]string) (*Response, error)
func (c *Client) Head(ctx context.Context, path string, headers map[string]string) (*Response, error)
func (c *Client) Post(ctx context.Context, path string, headers map[string]string, body []byte) (*Response, error)
func (c *Client) Put(ctx context.Context, path string, headers map[string]string, body []byte) (*Response, error)
func (c *Client) Delete(ctx context.Context, path string, headers map[string]string) (*Response, error)
func (c *Client) Patch(ctx context.Context, path string, headers map[string]string, body []byte) (*Response, error)
func (c *Client) Options(ctx context.Context, path string, headers map[string]string) (*Response, error)

// Convenience functions for common content types
func (c *Client) PostJSON(ctx context.Context, path string, headers map[string]string, data interface{}) (*Response, error)
func (c *Client) PutJSON(ctx context.Context, path string, headers map[string]string, data interface{}) (*Response, error)
func (c *Client) PatchJSON(ctx context.Context, path string, headers map[string]string, data interface{}) (*Response, error)
```

### Implementation Strategy

1. **Add Response Struct:** Define new `Response` type in `client.go`
2. **Add HTTP Method Functions:** Implement each HTTP method using consistent patterns
3. **Add Helper Functions:** Create internal helpers for request building and response processing
4. **Add Convenience Functions:** Provide JSON helpers for common use cases
5. **Maintain Compatibility:** Keep all existing functions unchanged

### Error Handling Strategy

- **Network Errors:** Return wrapped errors with context
- **HTTP Errors:** Return custom error type with status code and response body
- **JSON Errors:** Return serialization/deserialization errors for JSON functions
- **Context Errors:** Properly propagate context cancellation and timeout errors

## Acceptance Criteria

### AC-1: HTTP Method Implementation
- [ ] All 7 HTTP methods (GET, HEAD, POST, PUT, DELETE, PATCH, OPTIONS) implemented
- [ ] Functions accept correct parameters (context, path, headers, body when needed)
- [ ] Functions return `(*Response, error)` 

### AC-2: Response Structure
- [ ] Response struct contains StatusCode, Headers, and Body fields
- [ ] Headers are properly parsed from HTTP response
- [ ] Body is read completely and available as []byte

### AC-3: Context Support
- [ ] All functions accept context.Context as first parameter
- [ ] Context cancellation properly interrupts requests
- [ ] Context timeout respected

### AC-4: URL Handling
- [ ] Relative paths resolve against client base URL
- [ ] Absolute URLs override base URL  
- [ ] Special characters in paths properly encoded
- [ ] Query parameters in path preserved

### AC-5: Header Support
- [ ] Request headers properly set from map[string]string parameter
- [ ] Response headers captured in Response.Headers
- [ ] Content-Type automatically set for JSON convenience functions

### AC-6: Error Handling
- [ ] Network errors return descriptive wrapped errors
- [ ] HTTP error status codes (4xx, 5xx) return errors with status and body
- [ ] JSON serialization errors properly handled
- [ ] Context cancellation returns appropriate error

### AC-7: Backward Compatibility
- [ ] All existing scenario management functions continue working
- [ ] No breaking changes to existing Client struct or functions
- [ ] Existing tests continue passing

### AC-8: Testing
- [ ] Unit tests for all new HTTP method functions
- [ ] Tests cover success cases, error cases, and edge cases
- [ ] Tests verify context cancellation behavior
- [ ] Tests verify URL construction logic
- [ ] Test coverage maintains or improves current levels

### AC-9: Documentation
- [ ] Godoc documentation for all new public functions
- [ ] Usage examples in documentation
- [ ] README updated with client library examples

### AC-10: JSON Convenience Functions
- [ ] PostJSON, PutJSON, PatchJSON functions implemented
- [ ] Functions accept interface{} and serialize to JSON
- [ ] Content-Type: application/json automatically set
- [ ] JSON serialization errors properly handled

## Implementation Plan

### Phase 1: Core HTTP Methods (1-2 days)
1. Add Response struct definition
2. Implement basic HTTP method functions (GET, POST, PUT, DELETE, HEAD)
3. Add internal helper functions for request building and response processing
4. Add basic unit tests

### Phase 2: Extended Methods and Convenience Functions (1 day)
1. Implement PATCH and OPTIONS methods
2. Add JSON convenience functions (PostJSON, PutJSON, PatchJSON)
3. Enhance error handling with custom error types
4. Add comprehensive unit tests

### Phase 3: Testing and Documentation (1 day)
1. Add E2E tests using client against live Unimock server
2. Enhance unit test coverage for edge cases
3. Add godoc documentation
4. Update README with usage examples

### Phase 4: Integration and Validation (1 day)
1. Run full test suite to ensure backward compatibility
2. Performance testing to verify no significant overhead
3. Code review and refinements
4. Documentation review

## Success Metrics

- **Functional:** All 10 acceptance criteria met
- **Quality:** Test coverage ≥ 90% for new client functions
- **Performance:** <5ms overhead compared to raw http.Client for simple requests
- **Compatibility:** Zero breaking changes, all existing tests pass

## Dependencies

- Go standard library (`net/http`, `context`, `encoding/json`)
- Existing Unimock client package structure
- Existing test infrastructure

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|---------|------------|
| Breaking changes to existing client | High | Comprehensive testing, careful API design |
| Performance degradation | Medium | Performance testing, minimal abstraction overhead |
| Complex URL handling edge cases | Medium | Thorough testing of URL construction logic |
| Context handling inconsistencies | Medium | Follow Go best practices, test cancellation scenarios |

## Future Considerations

- **Authentication Support:** Add OAuth, API key, or basic auth support
- **Retry Logic:** Add configurable retry mechanisms for failed requests  
- **Response Streaming:** Support streaming large response bodies
- **Request/Response Middleware:** Allow pluggable request/response transformations
- **Mock Verification:** Add functions to verify expected requests were made