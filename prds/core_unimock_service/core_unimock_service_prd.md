# PRD: Core Unimock Service

## 1. Introduction

Unimock is a versatile mock service designed to handle HTTP requests and deliver mock responses based on highly configurable scenarios. This document outlines the core functional and non-functional requirements for the Unimock service, covering its fundamental request handling, resource management, content type support, configuration, error handling, and other essential operational aspects.

## 2. Goals

*   Establish a reliable and flexible mock service.
*   Support a wide range of HTTP interactions and data types.
*   Provide robust configuration options for tailoring mock behavior.
*   Ensure consistent performance and security.

## 3. User Stories

*   As a developer, I want to mock various HTTP GET requests so that I can test my client application's retrieval logic.
*   As a developer, I want to mock HTTP POST requests so that I can test resource creation flows.
*   As a developer, I want to mock HTTP PUT requests so that I can test resource update mechanisms.
*   As a developer, I want to mock HTTP DELETE requests so that I can test resource deletion processes.
*   As a QA engineer, I want to configure mock responses with various content types so that I can validate my application's handling of diverse data.
*   As a developer, I want to easily configure ID extraction rules so that Unimock can identify resources correctly.
*   As a system administrator, I want clear error messages so that I can diagnose issues effectively.

## 4. Requirements

### 4.1. General Request Handling
*   Must handle standard HTTP methods (GET, POST, PUT, DELETE).
*   Must support both individual resource and collection endpoints.
*   Must validate request content types where necessary (e.g., for body ID extraction).
*   Must handle non-existent resources appropriately (e.g., 404 Not Found).
*   Must support path-based routing.

### 4.2. Basic Resource Management (excluding advanced multi-ID features)

#### 4.2.1. GET Requests
*   Must first try to get resource by ID (last part of path).
*   If resource not found by ID, must try to get resources collection by the exact path (e.g., /users/999).
*   Must NOT fall back to parent collection (e.g., /users) if neither resource nor collection at the requested path exists.
*   Must return collection responses as a JSON array, including only resources with content type containing 'json'. (Note: This might conflict with "Response Formatting" section, review needed)
*   Must return 404 Not Found if neither resource nor collection is found.
*   Must return proper Content-Type headers.
*   Must sort collection responses by path for consistent ordering.
*   Must handle malformed collection paths.
*   Must support case-sensitive path matching (unless overridden by configuration).
*   Must handle trailing slashes in paths.
*   Must handle mixed content types in collection responses:
    *   Must preserve original content for all items (including JSON and binary data).
    *   Must return collection responses as application/json.
    *   Must handle future XML content type support.

#### 4.2.2. POST Requests
*   Must return 201 Created for new resources, including when an ID is auto-generated.
*   If no ID is extractable from the request (header, body, or path as per basic config), a new UUID must be generated and used as the resource ID.
*   Must validate request body if ID extraction from body is configured.
*   Must support ID extraction from basic sources (simple path in body, header, last path segment).
*   Must handle duplicate resource creation attempts (e.g., return 409 Conflict or as configured).
*   Must support Location header for created resources, pointing to `/path/to/collection/<id>`.
*   Must handle empty request bodies appropriately (e.g., accept if ID from path/header, or reject if body ID needed).
*   Must handle malformed JSON/XML if ID extraction from the body is configured and fails due to parsing errors (return 400 Bad Request).
*   Must handle missing required fields (if such validation is implemented).
*   Must handle invalid content types if relevant to processing.

#### 4.2.3. PUT Requests
*   Must first try to update resource by ID (last part of path).
*   Must return 404 Not Found if resource not found.
*   Must validate request body if configured.
*   Must maintain resource consistency.
*   Must handle partial updates (if supported).
*   Must handle empty request bodies.
*   Must handle malformed JSON/XML if relevant to processing.
*   Must handle missing required fields (if such validation is implemented).
*   Must handle invalid content types if relevant to processing.

#### 4.2.4. DELETE Requests
*   Must first try to delete resource by ID (last part of path).
*   If resource not found by ID, must try to delete all resources by the exact path (e.g., /users/999).
*   Must NOT fall back to parent collection (e.g., /users) if neither resource nor collection at the requested path exists.
*   Must return 404 Not Found if neither resource nor collection is found.
*   Must handle path-based deletion.
*   Must handle recursive deletion (if supported for collections).
*   Must handle non-existent paths.
*   Must handle malformed paths.
*   Must handle case-sensitive path matching (unless overridden by configuration).

### 4.3. Basic ID Extraction
(Details for simple ID extraction, complex/multi-ID in a separate PRD)

#### 4.3.1. GET Requests
*   Must extract ID from last path segment.
*   Must handle numeric IDs.
*   Must handle UUIDs.
*   Must handle special characters in IDs (ensure proper URL decoding).
*   Must handle missing IDs in collection paths (treat as collection request).
*   Must handle malformed path segments.

#### 4.3.2. POST Requests
*   Must extract ID from:
    *   JSON/XML body (simple path: `/id` - configurable).
    *   HTTP headers (configurable names).
*   If no ID is found via configured methods, a new UUID is generated and used for resource creation.
*   Must handle duplicate IDs (e.g., if trying to POST with an existing ID).
*   Must handle invalid ID formats (if validation is in place).
*   Must handle case-sensitive ID matching (based on configuration).

#### 4.3.3. PUT Requests
*   Must extract ID from last path segment.
*   May support ID from JSON/XML body (if different from path, configurable behavior for mismatch).
*   May support ID from HTTP headers (if configured).
*   Must handle ID mismatch between path and body (e.g., prefer path ID, or reject).
*   Must handle missing IDs (ID from path is primary).
*   Must handle invalid ID formats (if validation is in place).
*   Must handle case-sensitive ID matching (based on configuration).

#### 4.3.4. DELETE Requests
*   Must extract ID from last path segment.
*   May support ID from HTTP headers (if configured, for specific use cases).
*   Must handle missing IDs (ID from path is primary).
*   Must handle invalid ID formats (if validation is in place).
*   Must handle case-sensitive ID matching (based on configuration).

### 4.4. Content Type Support
*   Must support any content type for requests and responses.
*   Must return the original Content-Type for single resource responses.
*   Must return application/json for collection responses.
*   Must not validate or restrict content types by default (unless specific features require it).
*   Must handle content type with charset and parameters.

### 4.5. Configuration
*   Must support configuration for:
    *   Path patterns for routing/sectioning mocks.
    *   Basic ID extraction paths (e.g., simple body path, header name).
    *   Section-specific settings.
    *   Case sensitivity settings for paths and IDs.
    *   Default content types.
    *   Error response formats.
    *   Custom status codes for standard operations.
    *   Response headers (default or per-mock).
*   Configuration examples should be provided.

### 4.6. Error Handling
*   Must handle:
    *   Invalid requests (e.g., malformed HTTP).
    *   Resource not found scenarios (404).
    *   Duplicate resource creation (e.g., 409).
    *   Invalid content types if parsing is strictly required.
    *   Invalid JSON/XML bodies if parsing is strictly required for ID extraction.
    *   Missing required fields (if defined by specific features).
    *   Malformed paths.
    *   Invalid HTTP methods for a path (e.g., 405 Method Not Allowed).
    *   Missing headers if critical for operation.
    *   Invalid header values if critical.
    *   Timeout scenarios (if applicable).
    *   Concurrent modification conflicts (if applicable, e.g., optimistic locking).
    *   Storage errors (graceful degradation or clear error reporting).
    *   Configuration errors (startup or reload).

### 4.7. Storage Requirements
*   Must support concurrent access to storage.
*   Must maintain data consistency.
*   Must support path-based retrieval.
*   Must handle multiple resources per path (collections).
*   Must support efficient deletion of resources.
*   Must handle storage capacity limits gracefully.
*   Must handle storage errors gracefully.
*   Must support atomic operations where necessary (e.g., create/update).
*   Must maintain data integrity.
*   Must handle storage initialization errors.

### 4.8. Response Formatting
*   Must return collection responses as a JSON array.
*   Must not marshal or format individual items in the array (just concatenate as raw bytes, e.g. [item1, item2]) for non-JSON items. JSON items should be valid JSON within the array.
*   Must preserve original content for all items (including JSON, XML, binary, text, etc.).
*   Must return all types of resources as a slice of bytes without marshalling or formatting if they are not JSON.

### 4.9. Security Considerations
*   Must validate input data to prevent common vulnerabilities.
*   Must handle malformed requests gracefully to avoid crashes or exploits.
*   Must prevent resource leakage between tenants/contexts (if multi-tenancy is a feature).
*   Must validate content types if specific parsing is assumed.
*   Must sanitize response data if it includes user-provided content that could lead to XSS (less common for a mock service, but consider).
*   Must handle request size limits.
*   Must prevent path traversal vulnerabilities.
*   Must handle injection attempts (e.g., in configuration or dynamic parameters).
*   Must validate header values if they control critical logic.
*   Must handle authentication/authorization headers if such features are added.
*   Must prevent sensitive data exposure in errors or logs.

### 4.10. Performance Requirements
*   Must handle concurrent requests effectively.
*   Must maintain consistent response times under normal load.
*   Must efficiently handle large collections (within reasonable limits).
*   Must optimize storage operations.
*   Must support caching where appropriate (if applicable).
*   Must handle high request rates (define targets if possible).
*   Must handle large request/response bodies (define limits if possible).
*   Must handle connection limits gracefully.
*   Must handle resource limits (e.g., max number of mocks).
*   Must handle timeout scenarios gracefully.
*   Must support rate limiting (if applicable).

### 4.11. Basic Scenario Handling
*   Scenarios must be matched by RequestPath in the mock handler.
*   If a scenario is found by RequestPath, the mock handler must return the scenario details and skip all other mock handling logic.
    *   This implies scenarios can override status code, headers, and body.

## 5. Non-Functional Requirements

*   **Configurability:** The system must be highly configurable.
*   **Extensibility:** The design should allow for future extensions (e.g., new ID extraction methods, support for more protocols).
*   **Usability:** Easy to set up, configure, and run.
*   **Reliability:** The service should be stable and operate reliably.
*   **Maintainability:** Code should be well-structured and documented.

## 6. Out of Scope / Future Considerations

*   Advanced multi-ID resource identification (covered in a separate PRD).
*   Complex E2E scenario features (covered in a separate PRD).
*   Graphical User Interface for configuration.
*   Real-time synchronization of mock data across multiple instances (unless basic storage provides this).
*   Support for protocols other than HTTP/S.

## 7. Acceptance Criteria (Examples - to be detailed per requirement)

*   Given a GET request to `/users/123`, when user 123 exists, then the service returns a 200 OK with user 123 data.
*   Given a POST request to `/products` with valid product data, when the resource is created, then the service returns a 201 Created with a Location header.
*   Given a configuration for ID extraction from header `X-Transaction-ID`, when a POST request includes this header, then the resource is created using the header value as its ID.
---
*Self-Refinement: Added more structure based on example PRD, populated sections from docs/requirements.md. Some requirements (like collection response formatting) might need further review for consistency across sections.* 
