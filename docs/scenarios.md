# Test Scenarios for Unimock Requirements

This document lists test scenarios corresponding to the functional and non-functional requirements defined in `docs/requirements.md`. Each scenario should be covered by an End-to-End (E2E) test.

## Scenario Format

Each scenario should be described with enough detail to understand its purpose, steps, and expected outcome. Refer to specific requirement IDs from `docs/requirements.md`.

**Requirement Ref:** [Reference to heading/bullet in `docs/requirements.md`, e.g., REQ1.1 for 1st bullet under "1. Request Handling"]
**Scenario ID:** [Unique ID for this scenario, e.g., SCEN-REQ1.1-001]
**Description:** [Brief description of the test case, including common or edge cases.]
**Preconditions:** [Any setup needed before the test can run.]
**Steps:**
1. [Step 1]
2. [Step 2]
...
**Expected Result:** [What the outcome should be.]
**E2E Test Link/Reference:** [Link or reference to the E2E test implementing this scenario. TBD initially.]

---

## Scenarios

### Section 1: Request Handling (Corresponds to "### 1. Request Handling" in `docs/requirements.md`)

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must handle standard HTTP methods (GET, POST, PUT, DELETE)"
**Scenario ID:** SCEN-RH-001
**Description:** Verify successful processing of a GET request for an existing individual resource.
**Preconditions:**
    - Unimock service is running.
    - A mock resource is configured at `/test/resource/item123` with content `{"id": "item123", "data": "sample data"}` and Content-Type `application/json`.
**Steps:**
1. Send a GET request to `/test/resource/item123`.
**Expected Result:**
    - HTTP status code 200 OK.
    - Response body is `{"id": "item123", "data": "sample data"}`.
    - Content-Type header is `application/json`.
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must handle standard HTTP methods (GET, POST, PUT, DELETE)"
**Scenario ID:** SCEN-RH-002
**Description:** Verify successful processing of a POST request to create a new resource.
**Preconditions:**
    - Unimock service is running.
    - Unimock is configured to allow POST requests to `/test/collection` and extract ID from body path `/id`.
**Steps:**
1. Send a POST request to `/test/collection` with body `{"id": "newItem", "value": "new data"}` and Content-Type `application/json`.
**Expected Result:**
    - HTTP status code 201 Created.
    - Location header is present and points to `/test/collection/newItem`.
    - The resource `{"id": "newItem", "value": "new data"}` is stored and retrievable via GET `/test/collection/newItem`.
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must handle standard HTTP methods (GET, POST, PUT, DELETE)"
**Scenario ID:** SCEN-RH-003
**Description:** Verify successful processing of a PUT request to update an existing resource.
**Preconditions:**
    - Unimock service is running.
    - A mock resource `{"id": "existingItem", "value": "old data"}` exists at `/test/collection/existingItem`.
    - Unimock is configured to allow PUT requests to `/test/collection/{id}`.
**Steps:**
1. Send a PUT request to `/test/collection/existingItem` with body `{"id": "existingItem", "value": "updated data"}` and Content-Type `application/json`.
**Expected Result:**
    - HTTP status code 200 OK (or 204 No Content, depending on implementation choice for PUT).
    - The resource at `/test/collection/existingItem` is updated to `{"id": "existingItem", "value": "updated data"}`.
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must handle standard HTTP methods (GET, POST, PUT, DELETE)"
**Scenario ID:** SCEN-RH-004
**Description:** Verify successful processing of a DELETE request for an existing individual resource.
**Preconditions:**
    - Unimock service is running.
    - A mock resource exists at `/test/resource/itemToDelete`.
**Steps:**
1. Send a DELETE request to `/test/resource/itemToDelete`.
**Expected Result:**
    - HTTP status code 200 OK or 204 No Content.
    - Subsequent GET request to `/test/resource/itemToDelete` returns 404 Not Found.
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must support both individual resource and collection endpoints"
**Scenario ID:** SCEN-RH-005
**Description:** Verify GET request for an individual resource endpoint.
**Preconditions:**
    - Unimock service is running.
    - A mock resource is configured at `/individual/item001`.
**Steps:**
1. Send a GET request to `/individual/item001`.
**Expected Result:**
    - Service returns 200 OK with the configured mock response for `/individual/item001`.
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must support both individual resource and collection endpoints"
**Scenario ID:** SCEN-RH-006
**Description:** Verify GET request for a collection endpoint.
**Preconditions:**
    - Unimock service is running.
    - Multiple mock resources are configured under `/collection/items/` (e.g., `/collection/items/1`, `/collection/items/2`).
**Steps:**
1. Send a GET request to `/collection/items`.
**Expected Result:**
    - Service returns 200 OK.
    - Response body is a JSON array containing representations of resources under `/collection/items/` (respecting rules from REQ2 for collection GETs).
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must validate request content types"
**Scenario ID:** SCEN-RH-007
**Description:** Verify service rejects a POST request with an unsupported or invalid Content-Type header when validation is expected.
**Preconditions:**
    - Unimock service is running.
    - A specific endpoint `/restricted_post` is configured to only accept `application/json`.
**Steps:**
1. Send a POST request to `/restricted_post` with Content-Type `application/xml` and some XML body.
**Expected Result:**
    - Service returns an HTTP 415 Unsupported Media Type status code (or 400 Bad Request if that's the chosen behavior for invalid content type).
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must handle non-existent resources appropriately"
**Scenario ID:** SCEN-RH-008
**Description:** Verify GET request for a non-existent individual resource returns 404.
**Preconditions:**
    - Unimock service is running.
    - No resource is configured at `/nonexistent/item`.
**Steps:**
1. Send a GET request to `/nonexistent/item`.
**Expected Result:**
    - Service returns HTTP 404 Not Found.
**E2E Test Link/Reference:** TBD

**Requirement Ref:** `docs/requirements.md` - "### 1. Request Handling" -> "Must support path-based routing"
**Scenario ID:** SCEN-RH-009
**Description:** Verify that requests to different paths are routed to different handlers/mock configurations.
**Preconditions:**
    - Unimock service is running.
    - Mock resource A is configured at `/path/A` returning "Response A".
    - Mock resource B is configured at `/path/B` returning "Response B".
**Steps:**
1. Send a GET request to `/path/A`.
2. Send a GET request to `/path/B`.
**Expected Result:**
1. First request returns "Response A".
2. Second request returns "Response B".
**E2E Test Link/Reference:** TBD 
