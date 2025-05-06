# HTTP Methods Support

Unimock supports all standard HTTP methods with specific behavior for each. This document details how each method is handled.

## GET Requests

GET requests are used to retrieve existing resources.

### Single Item Retrieval
- ID is extracted from the last segment of the path (e.g., `/users/123` → ID is "123")
- Returns the item with its original content type
- Location header contains the full resource path
- Returns 404 if resource not found

### Collection Retrieval
- When no ID is found (e.g., `/users`), returns all items stored at that path
- Always returns a JSON array format, regardless of individual item content types
- Returns an empty array if no items found
- Example: `/users` will return all user resources stored under that path

### Deep Path Retrieval
- Treats deep paths (e.g., `/users/123/orders/456`) as single resources
- Uses the last path segment as the ID for lookup
- Returns 404 if resource not found

## POST Requests

POST requests are used to create new resources.

### ID Extraction
POST requests extract IDs in the following order:
1. Headers (if configured in the matching section)
2. Request body (for JSON and XML):
   - For JSON: Must have ID in body
   - For XML: Uses body ID or falls back to last path segment
3. Path segment (for non-JSON/XML requests)

### Behavior
- Accepts any path structure (collection, resource, or deep path)
- Returns 409 if resource already exists (ID conflict)
- Returns 201 on successful creation with:
  - Location header containing the full resource path
  - Created resource in response body
- For JSON requests without ID, returns 400

## PUT Requests

PUT requests are used to update existing resources.

### ID Extraction
- Uses GET-style ID extraction (last path segment)
- No body parsing for ID extraction (consistent with GET)
- Example paths:
  - `/users/123` → updates resource with ID "123"
  - `/users/123/orders/456` → updates resource with ID "456"

### Behavior
- Returns 404 for non-existent resources
- Returns 200 on successful update
- Location header contains the full resource path
- Updated resource is returned in response body

## DELETE Requests

DELETE requests are used to remove resources.

### Two-Step Deletion Process
1. Try ID-based deletion (GET logic)
   - Example: `/users/123` → deletes resource with ID "123"
2. Fall back to path-based deletion if ID-based fails
   - Example: If ID not found, deletes all resources under `/users/123/*`

### Behavior
- Returns 204 on successful deletion (no response body)
- Returns 404 if no resources found to delete
- Location header contains the path of the deleted resource

### Example Paths
- `/users/123` → first tries to delete resource with ID "123", then falls back to deleting all resources under `/users/123/*`
- `/users/123/orders` → first tries to delete resource with ID "orders", then falls back to deleting all resources under `/users/123/orders/*` 
