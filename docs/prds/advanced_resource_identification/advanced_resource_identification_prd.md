# PRD: Advanced Resource Identification

## 1. Introduction

This document details the requirements for an advanced resource identification feature in Unimock. The core idea is to allow a single resource instance to be identifiable and accessible via multiple unique identifiers. This enhances flexibility in how resources are referenced and managed, particularly in systems where entities naturally have several keys (e.g., an internal ID, a user-facing ID, a correlation ID from an external system).

This feature builds upon the basic ID extraction capabilities of Unimock.

## 2. Goals

*   Enable a single resource to be uniquely identified by more than one key.
*   Allow configuration of how these multiple identifiers are extracted during resource creation.
*   Ensure that all associated identifiers for a resource resolve to the same underlying data for all operations (GET, PUT, DELETE).
*   Maintain data integrity by preventing a single identifier from being ambiguously linked to multiple resources.

## 3. User Stories

*   As a developer, I want to create a resource and have Unimock automatically associate it with an ID from the `X-Correlation-ID` header and an `internalId` field from the request body, so that I can retrieve or update this resource using either of these identifiers later.
*   As a QA engineer, I want to test scenarios where a user is identified by `userId` in some requests and `sessionToken` in others, both pointing to the same user mock data, so I can validate complex user session flows.
*   As an API designer, I want to mock resources that have both a system-generated UUID and a human-readable SKU, so that the mock service mirrors the real API's flexibility.

## 4. Requirements (REQ-RM-MULTI-ID)

*   **R1: Multiple Identifiers per Resource:**
    *   A single resource instance MUST be identifiable and accessible via multiple unique identifiers.
*   **R2: Configurable Multi-ID Extraction at Creation:**
    *   During resource creation (e.g., POST), the system MUST be able to extract and associate multiple identifiers with the created resource.
    *   This includes, but is not limited to, identifiers from:
        *   HTTP headers (e.g., `X-User-ID`, `X-Correlation-ID`).
        *   Fields within the JSON/XML request body (e.g., `/id`, `/customIdField`, `//nestedId`).
    *   The specific headers and body paths (using a defined path syntax, e.g., JSONPath or a simplified version) for these additional identifiers MUST be configurable per Unimock section/path pattern.
    *   The primary ID (e.g., from URL path, or auto-generated if none from path) will also be one of these identifiers.
*   **R3: Consistent Resource Resolution:**
    *   All associated identifiers for a given resource MUST resolve to the same resource data for retrieval (GET), modification (PUT), and deletion (DELETE) operations.
    *   For example, if resource R was created with identifiers `id1` (path), `id2` (header), and `id3` (body), then GET `/path/id1`, GET `/path?someParam=id2`, or GET `/path?otherParam=id3` (assuming query mechanisms are in place to use secondary IDs) should all return resource R.
    *   Similarly, PUT/DELETE operations targeting any of these identifiers should affect resource R.
*   **R4: Identifier Uniqueness Constraint:**
    *   The system MUST ensure that any given identifier is not actively associated with more than one resource at any time to prevent ambiguity.
    *   If an attempt is made to create a new resource or update an existing one in a way that would assign an already-used secondary identifier (associated with a *different* resource), the operation MUST fail with an appropriate error (e.g., 409 Conflict).
*   **R5: Configuration:**
    *   The configuration mechanism (e.g., `config.yaml`) MUST allow specifying lists of header names and body paths for extracting these secondary identifiers, likely on a per-section basis.
    *   Example Configuration Snippet (conceptual):
      ```yaml
      sections:
        products:
          pathPattern: "/products/*"
          primaryID: "path"
          multiIdent:
            headerNames: ["X-Product-SKU", "X-External-Ref"]
            bodyIDPaths: ["/details/upc", "/internalCode"]
          # ... other configurations
      ```
*   **R6: Interaction with Existing ID Logic:**
    *   This feature should augment, not necessarily replace, existing ID extraction (from path, simple body field, header for primary ID).
    *   If a primary ID is determined (e.g., from the URL path like `/products/prod123`), this `prod123` is one of the identifiers. The multi-ID mechanism adds others to this set.
    *   If no primary ID is found via traditional means and one is auto-generated (e.g., UUID), this auto-generated ID becomes one of the identifiers.
*   **R7: Error Handling:**
    *   Appropriate error codes/messages must be returned for failures related to this feature (e.g., identifier conflict).

## 5. Non-Functional Requirements

*   **Performance:** The lookup of resources by any of their associated identifiers should be efficient.
*   **Scalability:** The system should handle a reasonable number of secondary identifiers per resource and a large number of resources with multiple identifiers without significant performance degradation.
*   **Clarity:** Configuration for this feature should be clear and understandable.

## 6. Out of Scope / Future Considerations

*   Complex query languages for retrieving resources by combinations of secondary identifiers (e.g., `sku=X AND brand=Y`). Initial scope is direct lookup by one of the configured secondary IDs.
*   Automatic synchronization or de-duplication if different resources are inadvertently created that *should* be the same based on secondary IDs (uniqueness constraint R4 aims to prevent this at creation/update time).
*   Mechanisms for listing or managing all secondary identifiers associated with a resource (beyond just using them for lookup).

## 7. Acceptance Criteria (Examples)

*   Given a POST to `/items` with header `X-Item-Code: ABC` and body `{"itemId": "123"}`,
    And the section is configured to use `X-Item-Code` and `/itemId` as secondary identifiers,
    When the resource is created (e.g., with path ID `xyz`),
    Then GET `/items/xyz` returns the item,
    And a mechanism to GET using `X-Item-Code: ABC` (if implemented, e.g. `/items?itemCode=ABC`) returns the same item,
    And a mechanism to GET using `itemId: 123` (if implemented, e.g. `/items?itemId=123`) returns the same item.
*   Given resource `R1` exists and is associated with secondary ID `sku: XYZ`,
    When an attempt is made to POST a new resource `R2` with `sku: XYZ` in its body (where SKU is a configured secondary ID),
    Then the request fails with a 409 Conflict (or similar) error.

---
*Self-Refinement: Focused this PRD on REQ-RM-MULTI-ID, structured it, added user stories, goals, and acceptance criteria examples.* 
