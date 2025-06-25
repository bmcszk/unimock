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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_001_GetExistingResource`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_002_PostCreateResource`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_003_PutUpdateResource`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_004_DeleteResource`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_005_GetIndividualResourceEndpoint`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_006_GetCollectionEndpoint`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_007_PostInvalidContentType`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_008_GetNonExistentResource`

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
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_RH_009_PathBasedRouting`

---

### Section 11: Scenario Handling (Corresponds to "### 11. Scenario Handling" in `docs/requirements.md`)

**Requirement Ref:** `docs/requirements.md` - "### 11. Scenario Handling" -> "Scenarios must be matched by RequestPath in the mock handler."
**Scenario ID:** SCEN-SH-001
**Description:** Verify that a configured scenario is matched by its exact RequestPath.
**Preconditions:**
    - Unimock service is running.
    - A scenario is configured with `RequestPath: "GET /custom/scenario/exact"`, `StatusCode: 200`, `ContentType: "application/json"`, `Data: "{\"message\": \"exact scenario matched\"}"`.
    - No other mock resource or conflicting scenario exists for this exact path and method.
**Steps:**
1. Send a GET request to `/custom/scenario/exact`.
**Expected Result:**
    - HTTP status code 200 OK.
    - Response body is `{"message": "exact scenario matched"}`.
    - Content-Type header is `application/json`.
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_SH_001_ExactPathScenarioMatch`

**Requirement Ref:** `docs/requirements.md` - "### 11. Scenario Handling" -> "Scenarios must be matched by RequestPath in the mock handler."
**Scenario ID:** SCEN-SH-002
**Description:** Verify that a configured scenario with a wildcard in RequestPath is matched.
**Preconditions:**
    - Unimock service is running.
    - A scenario is configured with `RequestPath: "POST /custom/scenario/*"`, `StatusCode: 201`, `ContentType: "text/plain"`, `Data: "wildcard scenario matched"`.
    - No other mock resource or conflicting scenario exists for this path pattern and method.
**Steps:**
1. Send a POST request to `/custom/scenario/anything/here` with any body.
**Expected Result:**
    - HTTP status code 201 Created.
    - Response body is `wildcard scenario matched`.
    - Content-Type header is `text/plain`.
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_SH_002_WildcardPathScenarioMatch`

**Requirement Ref:** `docs/requirements.md` - "### 11. Scenario Handling" -> "If a scenario is found by RequestPath, the mock handler must return the scenario details and skip all other mock handling logic."
**Scenario ID:** SCEN-SH-003
**Description:** Verify that if a scenario matches, normal mock resource handling for the same path is skipped.
**Preconditions:**
    - Unimock service is running.
    - A scenario is configured with `RequestPath: "GET /override/path"`, `StatusCode: 299`, `ContentType: "application/xml"`, `Data: "<scenario>overridden</scenario>"`.
    - A regular mock resource is also configured at `/override/path` (e.g., via POSTing to it, or static configuration) with different content, say `{"id": "original", "value": "this should be skipped"}` and Content-Type `application/json`.
**Steps:**
1. Send a GET request to `/override/path`.
**Expected Result:**
    - HTTP status code 299.
    - Response body is `<scenario>overridden</scenario>`.
    - Content-Type header is `application/xml`.
    - The regular mock resource data is NOT returned.
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_SH_003_ScenarioSkipsMockHandling`

**Requirement Ref:** `docs/requirements.md` - "### 11. Scenario Handling" -> "Scenarios must be matched by RequestPath in the mock handler."
**Scenario ID:** SCEN-SH-004
**Description:** Verify that a scenario for a specific HTTP method (e.g. PUT) does not match for a different method (e.g. GET) on the same path.
**Preconditions:**
    - Unimock service is running.
    - A scenario is configured with `RequestPath: "PUT /specific/method/test"`, `StatusCode: 200`, `ContentType: "application/json"`, `Data: "{\"message\": \"PUT scenario matched\"}"`.
    - A regular mock resource exists at `/specific/method/test` configured via POST: `{"id": "regular", "data": "GET response"}` (Content-Type `application/json`).
**Steps:**
1. Send a PUT request to `/specific/method/test` with body `{"data": "update"}`.
2. Send a GET request to `/specific/method/test`.
**Expected Result:**
1. For the PUT request:
    - HTTP status code 200 OK.
    - Response body is `{"message": "PUT scenario matched"}`.
    - Content-Type header is `application/json`.
2. For the GET request:
    - HTTP status code 200 OK.
    - Response body is `{"id": "regular", "data": "GET response"}`.
    - Content-Type header is `application/json`.
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_SH_004_ScenarioMethodMismatch`

**Requirement Ref:** `docs/requirements.md` - "### 11. Scenario Handling"
**Scenario ID:** SCEN-SH-005
**Description:** Verify scenario matching with an empty data field and custom location header.
**Preconditions:**
    - Unimock service is running.
    - A scenario is configured with `RequestPath: "POST /resource/creation"`, `StatusCode: 201`, `ContentType: "application/json"`, `Data: ""`, `Location: "/resource/creation/new-id-from-scenario"`.
**Steps:**
1. Send a POST request to `/resource/creation` with body `{"name": "test"}`.
**Expected Result:**
    - HTTP status code 201 Created.
    - Response body is empty.
    - Content-Type header is `application/json`.
    - Location header is `/resource/creation/new-id-from-scenario`.
**E2E Test Link/Reference:** `e2e/request_handling_test.go#TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation`

---

### Section X: Resource Management - Multiple IDs (REQ-RM-MULTI-ID)

**Requirement Ref:** `docs/requirements.md` - "### X. Resource Management" -> "REQ-RM-MULTI-ID: A single resource can be identified and manipulated using multiple external IDs."
**Scenario ID:** SCEN-RM-MULTI-ID-001
**Description:** Create a resource with multiple IDs (one from header, one from body JSON path) and verify it can be retrieved by either ID.
**Preconditions:**
    - Unimock service is running.
    - Mock config section `products` exists with `path_pattern: "/products/*"`, `header_id_name: "X-Product-Token"`, `body_id_paths: ["/product/sku"]`.
**Steps:**
1. Send a POST request to `/products` with header `X-Product-Token: token123` and JSON body `{"product": {"sku": "skuABC"}, "name": "Multi-ID Product"}`.
2. Expected: HTTP 201 Created. Location header might point to one of the IDs (e.g., `/products/token123` or `/products/skuABC`).
3. Send a GET request to `/products/token123`.
4. Expected: HTTP 200 OK. Response body is `{"product": {"sku": "skuABC"}, "name": "Multi-ID Product"}`.
5. Send a GET request to `/products/skuABC`.
6. Expected: HTTP 200 OK. Response body is `{"product": {"sku": "skuABC"}, "name": "Multi-ID Product"}`.
**E2E Test Link/Reference:** `e2e/e2e_test.go#TestE2E_SCEN_RM_MULTI_ID_001`

**Requirement Ref:** `docs/requirements.md` - "### X. Resource Management" -> "REQ-RM-MULTI-ID: A single resource can be identified and manipulated using multiple external IDs."
**Scenario ID:** SCEN-RM-MULTI-ID-002
**Description:** Update a resource (identified by one of its multiple IDs) and verify the update is reflected when retrieving by another of its IDs.
**Preconditions:**
    - Unimock service is running.
    - A resource exists, associated with external IDs `id_A` and `id_B`. Original data: `{"value": "original"}`.
    - Mock config allows PUT to `/items/{id}`.
**Steps:**
1. Send a PUT request to `/items/id_A` with JSON body `{"value": "updated"}`.
2. Expected: HTTP 200 OK (or 204).
3. Send a GET request to `/items/id_B`.
4. Expected: HTTP 200 OK. Response body is `{"value": "updated"}`.
**E2E Test Link/Reference:** `e2e/e2e_test.go#TestE2E_SCEN_RM_MULTI_ID_002`

**Requirement Ref:** `docs/requirements.md` - "### X. Resource Management" -> "REQ-RM-MULTI-ID: A single resource can be identified and manipulated using multiple external IDs."
**Scenario ID:** SCEN-RM-MULTI-ID-003
**Description:** Delete a resource (identified by one of its multiple IDs) and verify it's no longer accessible by any of its other associated IDs.
**Preconditions:**
    - Unimock service is running.
    - A resource exists, associated with external IDs `id_X`, `id_Y`, and `id_Z`.
    - Mock config allows DELETE to `/resources/{id}`.
**Steps:**
1. Send a DELETE request to `/resources/id_Y`.
2. Expected: HTTP 200 OK or 204 No Content.
3. Send a GET request to `/resources/id_X`.
4. Expected: HTTP 404 Not Found.
5. Send a GET request to `/resources/id_Z`.
6. Expected: HTTP 404 Not Found.
**E2E Test Link/Reference:** `e2e/e2e_test.go#TestE2E_SCEN_RM_MULTI_ID_003`

**Requirement Ref:** `docs/requirements.md` - "### X. Resource Management" -> "REQ-RM-MULTI-ID: A single resource can be identified and manipulated using multiple external IDs."
**Scenario ID:** SCEN-RM-MULTI-ID-004
**Description:** Attempt to create a new resource providing an external ID that is already associated with an existing resource, verify conflict.
**Preconditions:**
    - Unimock service is running.
    - A resource exists and is associated with external ID `existing_token`.
    - Mock config section `gadgets` exists with `path_pattern: "/gadgets/*"`, `header_id_name: "X-Gadget-Token"`.
**Steps:**
1. Send a POST request to `/gadgets` with header `X-Gadget-Token: existing_token` and JSON body `{"name": "Conflicting Gadget"}`.
**Expected Result:**
    - HTTP status code 409 Conflict.
    - The original resource associated with `existing_token` remains unchanged.
**E2E Test Link/Reference:** `e2e/e2e_test.go#TestE2E_SCEN_RM_MULTI_ID_004`

**Requirement Ref:** `docs/requirements.md` - "### X. Resource Management" -> "REQ-RM-MULTI-ID: A single resource can be identified and manipulated using multiple external IDs."
**Scenario ID:** SCEN-RM-MULTI-ID-005
**Description:** Create a resource via POST where IDs are extracted from multiple body paths (JSON), retrieve by each.
**Preconditions:**
    - Unimock service is running.
    - Mock config section `documents` exists with `path_pattern: "/documents/*"`, `body_id_paths: ["/meta/uuid", "/alt_id"]`.
**Steps:**
1. Send a POST request to `/documents` with JSON body `{"meta": {"uuid": "docUUID1"}, "alt_id": "altIDXYZ", "content": "Test document"}`.
2. Expected: HTTP 201 Created.
3. Send a GET request to `/documents/docUUID1`.
4. Expected: HTTP 200 OK with the document body.
5. Send a GET request to `/documents/altIDXYZ`.
6. Expected: HTTP 200 OK with the document body.
**E2E Test Link/Reference:** `e2e/e2e_test.go#TestE2E_SCEN_RM_MULTI_ID_005`

---

### Section 12: Complex End-to-End Scenarios (Corresponds to "### 12. Complex End-to-End Scenarios" in `docs/requirements.md`)

**Requirement Ref:** `docs/requirements.md` - "### 12. Complex End-to-End Scenarios" -> "REQ-E2E-COMPLEX-001: Multistage Resource Lifecycle with Scenario Override"
**Scenario ID:** SCEN-E2E-COMPLEX-001
**Description:** Verify the complete lifecycle of a resource, including creation with multiple identifiers, retrieval, updates, dynamic behavior modification via Unimock scenarios, and eventual deletion. This scenario tests the integration of resource management and scenario overriding capabilities.
**Preconditions:**
    - Unimock service is running.
    - Unimock is configured to support a `/products` collection.
    - ID extraction is configured for `POST /products`:
        - Primary ID: auto-generated UUID if not in path.
        - Secondary ID from body: e.g., `body_id_paths: ["/sku"]`.
    - Unimock scenario management endpoints are available (e.g., `POST /_unimock/scenarios`, `DELETE /_unimock/scenarios/{scenario_id}`).
**Steps:**
1.  **Create Resource:** Send a POST request to `/products` with body `{"sku": "E2ECOMPLEXSKU001", "name": "Complex Product", "version": 1}`.
    - Verify HTTP 201 Created.
    - Extract the auto-generated primary ID (e.g., `productID1`) from the Location header or response.
    - Store `productID1` and `sku: "E2ECOMPLEXSKU001"`.
2.  **Retrieve by Secondary ID (SKU):** Send a GET request to `/products?sku=E2ECOMPLEXSKU001`. (This step implies functionality to query by attributes. If not directly supported, this step might need to be adapted, e.g. GET /products/E2ECOMPLEXSKU001 if SKU is treated as an ID).
    - Verify HTTP 200 OK.
    - Verify response body contains `{"sku": "E2ECOMPLEXSKU001", "name": "Complex Product", "version": 1}`.
3.  **Update Resource:** Send a PUT request to `/products/{productID1}` with body `{"sku": "E2ECOMPLEXSKU001", "name": "Complex Product Updated", "version": 2}`.
    - Verify HTTP 200 OK (or 204 No Content).
4.  **Verify Update:** Send a GET request to `/products/{productID1}`.
    - Verify HTTP 200 OK.
    - Verify response body contains `{"sku": "E2ECOMPLEXSKU001", "name": "Complex Product Updated", "version": 2}`.
5.  **Apply Unimock Scenario:** Create a Unimock scenario targeting `GET /products/{productID1}`.
    - Request: POST to `/_unimock/scenarios` with body like:
      `{"request_path": "GET /products/{productID1}", "status_code": 418, "response_body": "{\"message\": \"I am a teapot - scenario active!\"}", "content_type": "application/json"}`.
    - Verify HTTP 201 Created (for scenario creation).
    - Store the scenario ID (e.g., `scenarioID1`).
6.  **Verify Scenario Active:** Send a GET request to `/products/{productID1}`.
    - Verify HTTP 418 I'm a teapot.
    - Verify response body is `{"message": "I am a teapot - scenario active!"}`.
7.  **Delete Unimock Scenario:** Send a DELETE request to `/_unimock/scenarios/{scenarioID1}`.
    - Verify HTTP 200 OK or 204 No Content (for scenario deletion).
8.  **Verify Scenario Removal (Resource Reverts):** Send a GET request to `/products/{productID1}`.
    - Verify HTTP 200 OK.
    - Verify response body contains `{"sku": "E2ECOMPLEXSKU001", "name": "Complex Product Updated", "version": 2}` (back to actual data).
9.  **Delete Resource:** Send a DELETE request to `/products/{productID1}`.
    - Verify HTTP 200 OK or 204 No Content.
10. **Verify Deletion:** Send a GET request to `/products/{productID1}`.
    - Verify HTTP 404 Not Found.
**Expected Result:** All steps complete successfully, verifying the interactions between resource lifecycle and scenario management.
**E2E Test Link/Reference:** e2e/e2e_complex_lifecycle_test.go#TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle (Covered by TASK-030)

---