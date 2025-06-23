# PRD: Request/Response Transformation

## 1. Introduction

This PRD defines the requirements for implementing optional Go function-based request and response transformation capabilities in the Unimock service. This feature enables users to programmatically modify incoming requests and outgoing responses when using Unimock as a library, providing flexibility for advanced testing scenarios and API behavior simulation.

## 2. Goals

* Enable programmatic transformation of HTTP requests before they reach the mock handler
* Enable programmatic transformation of HTTP responses before they are returned to clients
* Provide a clean, Go function-based API for transformation logic
* Support configuration through Go code when using Unimock as a library
* Maintain backward compatibility with existing library usage patterns
* Ensure transformation logic is only available in library mode (not via YAML configuration)

## 3. User Stories

* As a Go developer using Unimock as a library, I want to transform incoming requests so that I can simulate different API behaviors and test edge cases
* As a Go developer, I want to transform outgoing responses so that I can inject test-specific headers, modify response bodies, or simulate error conditions
* As a library user, I want to configure transformations per endpoint section so that I can apply different logic to different API paths
* As a developer, I want transformation functions to have access to request context so that I can implement stateful transformation logic
* As a library user, I want transformations to be optional so that I can use Unimock without any transformation overhead when not needed

## 4. Requirements

### 4.1. Core Transformation Framework

#### 4.1.1. Function Signatures
* Must define clear Go function signatures for request and response transformations
* Must support transformation functions that can return errors for validation/rejection scenarios
* Must provide access to the original request/response objects and allow modifications
* Must support context passing for stateful transformations

#### 4.1.2. Configuration API
* Must extend the library configuration API to accept transformation functions
* Must support per-section (endpoint pattern) transformation configuration
* Must NOT expose transformation functions in YAML configuration
* Must maintain backward compatibility with existing `NewServer()` API

#### 4.1.3. Integration Points
* Must integrate transformation hooks into the mock handler request processing flow
* Must apply request transformations before ID extraction and service layer processing
* Must apply response transformations after service layer processing but before final response
* Must handle transformation errors gracefully with appropriate HTTP error responses

### 4.2. Request Transformation

#### 4.2.1. Request Transformation Capabilities
* Must allow modification of HTTP method, URL path, query parameters
* Must allow modification of HTTP headers (add, update, remove)
* Must allow modification of request body content
* Must allow complete request rejection with custom error responses
* Must preserve request context throughout transformation

#### 4.2.2. Request Transformation Timing
* Must execute request transformations after routing but before ID extraction
* Must execute transformations before mock service operations
* Must support multiple transformation functions per section (chain execution)
* Must stop processing on first transformation error

### 4.3. Response Transformation

#### 4.3.1. Response Transformation Capabilities
* Must allow modification of HTTP status code
* Must allow modification of response headers (add, update, remove)
* Must allow modification of response body content
* Must allow complete response replacement
* Must preserve response context throughout transformation

#### 4.3.2. Response Transformation Timing
* Must execute response transformations after mock service operations
* Must execute transformations before final HTTP response writing
* Must support multiple transformation functions per section (chain execution)
* Must handle transformation errors by returning 500 Internal Server Error

### 4.4. Configuration Structure

#### 4.4.1. Library Configuration Extension
* Must extend existing configuration structures to include transformation functions
* Must support optional transformation configuration (nil-safe)
* Must provide clear separation between YAML-configurable and code-only options
* Must maintain existing configuration loading mechanisms

#### 4.4.2. Per-Section Configuration
* Must allow transformation configuration per mock section
* Must support different transformation logic for different endpoint patterns
* Must inherit global transformation settings when section-specific not provided
* Must validate transformation function configuration at service startup

### 4.5. Error Handling

#### 4.5.1. Transformation Error Handling
* Must handle transformation function panics gracefully
* Must provide clear error messages for transformation failures
* Must support custom error responses from transformation functions
* Must log transformation errors with appropriate detail level

#### 4.5.2. Fallback Behavior
* Must define clear fallback behavior when transformations fail
* Must support graceful degradation options (skip transformation vs. fail request)
* Must handle nil transformation functions safely
* Must validate transformation function configurations

### 4.6. Performance Considerations

#### 4.6.1. Transformation Overhead
* Must minimize performance impact when transformations are not configured
* Must support efficient transformation function execution
* Must handle transformation function timeouts if needed
* Must avoid memory leaks in transformation processing

#### 4.6.2. Concurrency Support
* Must support concurrent transformation execution
* Must ensure thread-safety in transformation function calls
* Must handle request/response object mutations safely
* Must support stateless and stateful transformation patterns

### 4.7. Security Considerations

#### 4.7.1. Transformation Security
* Must validate transformation function inputs
* Must prevent transformation functions from accessing unauthorized resources
* Must handle malicious transformation logic gracefully
* Must ensure transformation functions cannot break service stability

#### 4.7.2. Data Protection
* Must ensure sensitive data is not exposed through transformation errors
* Must handle request/response data modifications securely
* Must validate transformed content for injection vulnerabilities
* Must maintain audit trail for transformation activities

## 5. Non-Functional Requirements

* **Flexibility:** Support diverse transformation use cases without imposing rigid patterns
* **Performance:** Minimal overhead when transformations are not used
* **Reliability:** Transformation failures should not crash the service
* **Maintainability:** Clear separation between core mock functionality and transformation features
* **Testability:** Transformation functions should be easily unit testable

## 6. Out of Scope

* YAML-based transformation configuration (code-only feature)
* GUI for transformation configuration
* Predefined transformation templates or libraries
* Automatic transformation discovery or registration
* Transformation function hot-reloading
* Complex transformation orchestration or workflows

## 7. API Design (Preliminary)

### 7.1. Function Signatures

```go
// RequestTransformFunc transforms an HTTP request
type RequestTransformFunc func(ctx context.Context, req *http.Request, section string) (*http.Request, error)

// ResponseTransformFunc transforms an HTTP response
type ResponseTransformFunc func(ctx context.Context, req *http.Request, resp *http.Response, section string) (*http.Response, error)
```

### 7.2. Configuration Structure

```go
// TransformationConfig holds transformation functions for a section
type TransformationConfig struct {
    RequestTransforms  []RequestTransformFunc
    ResponseTransforms []ResponseTransformFunc
    FailureMode        TransformFailureMode // Continue, Abort, etc.
}

// Enhanced Section configuration
type Section struct {
    // Existing fields...
    PathPattern     string `yaml:"path_pattern"`
    BodyIDPaths     []string `yaml:"body_id_paths"`
    HeaderIDName    string `yaml:"header_id_name"`
    
    // New transformation field (not in YAML)
    Transformations *TransformationConfig `yaml:"-"`
}
```

### 7.3. Library Usage

```go
// Enhanced server creation with transformations
serverConfig := config.FromEnv()
mockConfig, err := config.LoadFromYAML("config.yaml")

// Add transformation functions
mockConfig.Sections["users"].Transformations = &config.TransformationConfig{
    RequestTransforms: []config.RequestTransformFunc{
        func(ctx context.Context, req *http.Request, section string) (*http.Request, error) {
            // Add custom header
            req.Header.Set("X-Test-Mode", "true")
            return req, nil
        },
    },
    ResponseTransforms: []config.ResponseTransformFunc{
        func(ctx context.Context, req *http.Request, resp *http.Response, section string) (*http.Response, error) {
            // Add response header
            resp.Header.Set("X-Transformed", "true")
            return resp, nil
        },
    },
}

srv, err := pkg.NewServer(serverConfig, mockConfig)
```

## 8. Acceptance Criteria

* Given a library user configures request transformation functions, when requests are processed, then the transformation functions are executed before ID extraction
* Given a library user configures response transformation functions, when responses are generated, then the transformation functions are executed before response delivery
* Given transformation functions return errors, when processing requests, then appropriate HTTP error responses are returned
* Given no transformation functions are configured, when processing requests, then performance overhead is minimal
* Given YAML configuration is loaded, when transformation functions are added via code, then both configurations work together seamlessly

---
*Created: 2025-06-23 - Initial PRD for request/response transformation feature in library mode*