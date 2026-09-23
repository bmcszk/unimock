# PRD: Strict Path Matching and Wildcard Pattern Support

## Overview

This PRD defines the implementation of strict path matching and enhanced wildcard pattern support for Unimock's configuration system. This feature improves the precision of path matching for GET, PUT, and DELETE operations while maintaining backward compatibility.

## Requirements

### REQ-PATH-001: Strict Path Flag
- **Description**: Add `strict_path` boolean flag to section configuration (default: false)
- **Behavior**: When enabled, GET/DELETE/UPDATE operations must match exact path/ID to elements, otherwise return 404
- **Default**: false (maintains current behavior for backward compatibility)

### REQ-PATH-002: Enhanced Wildcard Patterns
- **Description**: Extend `path_pattern` to support standard wildcard notation
- **Wildcards**:
  - `*` - Matches single path segment (current behavior)
  - `**` - Matches multiple path segments recursively
- **Examples**:
  - `/api/users/*` - matches `/api/users/123` but not `/api/users/123/posts`
  - `/api/**` - matches `/api/users/123/posts/456/comments`
  - `/api/*/posts/*` - matches `/api/users/posts/123`

### REQ-PATH-003: Strict Matching Logic
- **When strict_path=true**:
  - More restrictive path pattern matching rules apply
  - PUT `/users/123` only succeeds if resource with exact ID "123" exists (no upsert)
  - Stricter validation of path patterns against requests
- **When strict_path=false** (default):
  - More flexible path pattern matching (e.g., `/users/*` matches both `/users/123` and `/users`)
  - PUT performs upsert operations (creates if doesn't exist)
  - Less restrictive path matching rules
- **Always (regardless of strict_path setting)**:
  - Individual resource requests (e.g., GET `/users/123`) return 404 if resource doesn't exist
  - Collection requests (e.g., GET `/users`) return the collection if any resources exist

### REQ-PATH-004: Configuration Schema
```yaml
sections:
  users:
    path_pattern: "/users/*"
    strict_path: false  # default - flexible matching
    body_id_paths: ["/id"]
  
  admin_users:
    path_pattern: "/admin/users/*" 
    strict_path: true   # strict matching - both path AND resource ID must exist
    body_id_paths: ["/id"]
    
  api_resources:
    path_pattern: "/api/**"  # recursive matching
    strict_path: false
    body_id_paths: ["/id"]
```

### Practical Example

Given this configuration:
```yaml
sections:
  regular_users:
    path_pattern: "/users/*"
    strict_path: false
    body_id_paths: ["/id"]
  
  admin_users:
    path_pattern: "/admin/users/*"
    strict_path: true
    body_id_paths: ["/id"]
```

**Behavior when resource "123" exists and "999" doesn't exist:**

| Request | strict_path=false | strict_path=true |
|---------|------------------|------------------|
| `GET /users/123` | ✅ Returns resource | ✅ Returns resource |
| `GET /users/999` | ❌ Returns 404 | ❌ Returns 404 |
| `GET /users` | ✅ Returns collection | ✅ Returns collection |
| `PUT /users/999` | ✅ Creates new resource (upsert) | ❌ Returns 404 |
| `DELETE /users/999` | ❌ Returns 404 | ❌ Returns 404 |

**Key differences:**
- **Path Matching**: `strict_path=false` allows `/users/*` to match `/users` (collection), while `strict_path=true` requires exact pattern matching
- **PUT Behavior**: `strict_path=false` allows upsert (create if doesn't exist), while `strict_path=true` requires resource to exist first
- **Individual Resources**: Both settings return 404 for non-existent individual resources

## Implementation Tasks

### Task 1: Configuration Updates
- Add `StrictPath bool` field to `config.Section` struct
- Update YAML parsing to support `strict_path` field
- Ensure default value is `false` for backward compatibility

### Task 2: Wildcard Pattern Engine
- Implement enhanced pattern matching for `*` and `**` wildcards
- Update `MockConfig.MatchPath()` method
- Ensure backward compatibility with existing single `*` patterns

### Task 3: Handler Logic Updates
- Modify GET handler to respect strict path matching
- Modify PUT handler to disable upsert when strict_path=true
- Modify DELETE handler to enforce exact resource existence
- Maintain existing behavior when strict_path=false

### Task 4: Path Validation
- Implement exact ID matching validation
- Add path segment counting and validation
- Ensure proper 404 responses when strict matching fails

## Acceptance Criteria

### AC-001: Configuration Support
- [ ] `strict_path` field correctly parsed from YAML
- [ ] Default value is `false` when not specified
- [ ] Field accessible in handler logic

### AC-002: Wildcard Pattern Matching
- [ ] `*` matches single path segment
- [ ] `**` matches multiple path segments recursively
- [ ] Existing patterns continue to work
- [ ] Complex patterns like `/api/*/posts/**` work correctly

### AC-003: Strict Path Behavior
- [ ] GET with strict_path=true returns 404 for non-existent resources
- [ ] PUT with strict_path=true returns 404 for non-existent resources (no upsert)
- [ ] DELETE with strict_path=true returns 404 for non-existent resources
- [ ] All operations work normally when strict_path=false

### AC-004: Backward Compatibility
- [ ] Existing configurations work without changes
- [ ] Default behavior unchanged when strict_path not specified
- [ ] All existing tests continue to pass

## Test Cases

### Test Case 1: Strict Path Configuration
```yaml
# Configuration with strict path enabled
sections:
  strict_users:
    path_pattern: "/strict/users/*"
    strict_path: true
    body_id_paths: ["/id"]
```

**Test Steps:**
1. POST `/strict/users` with `{"id": "123", "name": "test"}`
2. GET `/strict/users/123` → should return 200 with resource
3. GET `/strict/users/999` → should return 404 (not fall back to collection)
4. PUT `/strict/users/999` with data → should return 404 (no upsert)
5. DELETE `/strict/users/999` → should return 404

### Test Case 2: Wildcard Patterns
```yaml
# Configuration with recursive wildcard
sections:
  api_resources:
    path_pattern: "/api/**"
    strict_path: false
    body_id_paths: ["/id"]
```

**Test Steps:**
1. POST `/api/users/123/posts` should match pattern
2. POST `/api/v1/admin/users/456/settings/profile` should match pattern
3. POST `/other/api/users` should NOT match pattern

### Test Case 3: Mixed Wildcards
```yaml
# Configuration with mixed wildcards
sections:
  user_posts:
    path_pattern: "/users/*/posts/*"
    strict_path: true
    body_id_paths: ["/id"]
```

**Test Steps:**
1. POST `/users/123/posts/456` should match
2. POST `/users/123/posts` should NOT match (missing second segment)
3. POST `/users/123/posts/456/comments` should NOT match (extra segment)

## Migration Strategy

1. **Phase 1**: Add configuration field with default `false`
2. **Phase 2**: Implement wildcard pattern engine
3. **Phase 3**: Update handler logic to respect strict_path flag
4. **Phase 4**: Add comprehensive tests
5. **Phase 5**: Update documentation

## Risk Assessment

- **Low Risk**: Configuration changes are additive with safe defaults
- **Medium Risk**: Pattern matching changes could affect existing complex patterns
- **Mitigation**: Comprehensive testing of existing patterns before release
- **Rollback**: Feature flag can be disabled by setting strict_path=false globally

## Success Metrics

- All existing tests pass without modification
- New strict path tests achieve 100% coverage
- Performance impact < 5% for pattern matching operations
- Zero breaking changes to existing API behavior