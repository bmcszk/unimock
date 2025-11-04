# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Reference

**Primary weapons - use them often:**
- `make check` - Your PRIMARY weapon (vet, lint, unit tests) - MUST pass before commits
- `make test-all` - Your SECONDARY weapon (unit + E2E tests) - MUST pass before push

**Git hooks enforce quality:**
- Pre-commit: Runs `make check` automatically
- Pre-push: Runs `make test-all` automatically

## Essential Commands

### Core Development
```bash
make build      # Build the unimock binary
make run        # Build and run locally
make test-unit  # Unit tests with race detection
make test-e2e   # E2E tests with actual HTTP server
make vet        # Go vet static analysis
make lint       # golangci-lint with strict rules
make tidy       # Clean up dependencies
```

### Run Specific Tests
```bash
go test ./pkg/config -v -run TestSpecificFunction
gotestsum -- ./pkg/config -v -run TestUniConfig_LoadFromYAML
```

## Architecture Overview

Unimock is a **universal HTTP mock server** with clean layered architecture:

```
HTTP Request → Chi Router → Handler → Service → Storage → Response
```

### Core Components

**Configuration System** (`pkg/config/`):
- Unified YAML format: `sections` (mock behavior) + `scenarios` (predefined responses)
- Legacy format compatibility
- Smart ID extraction from URLs, JSON/XML bodies, HTTP headers
- Wildcard pattern matching: `*` (single segment), `**` (recursive)
- Fixture file support with go-restclient compatible syntax

**Request Flow**:
1. **Router** (`internal/router/`) - Chi v5 with middleware, scenario override logic
2. **Handler** (`internal/handler/`) - UniHandler (CRUD), ScenarioHandler (overrides), TechHandler (health/metrics)
3. **Service** (`internal/service/`) - Business logic, conflict detection, scenario matching
4. **Storage** (`internal/storage/`) - Thread-safe in-memory storage with composite keys

**Key Concepts**:
- **Composite keys**: Resources identified by multiple IDs (path, body, headers)
- **Strict vs Flexible paths**: Enforce creation path structure or allow cross-path access
- **Conflict detection**: Prevent simultaneous resource modifications

## Configuration Patterns

### Section Configuration (Mock Behavior)
```yaml
sections:
  users:
    path_pattern: "/api/users/*"           # Wildcard: * (segment), ** (recursive)
    body_id_paths: ["/id", "/user/id"]    # JSON/XML extraction
    header_id_names: ["X-User-ID"]        # Multiple headers supported
    return_body: true                      # Response body control
    strict_path: false                    # Path access mode
```

### Scenario Overrides (Predefined Responses)
```yaml
scenarios:
  - uuid: "user-not-found"
    method: "GET"
    path: "/api/users/999"
    status_code: 404
    data: '{"error": "User not found"}'
```

### Fixture File Support (go-restclient Compatible)

**Three syntax styles:**
```yaml
# 1. Legacy @ syntax (backward compatibility)
data: "@fixtures/operations/robots.json"

# 2. Space-after-< syntax (SPACE REQUIRED after <)
data: "< ./fixtures/users/user.json"

# 3. <@ syntax (@ IMMEDIATELY after <, no space between < and @)
data: "<@ ./fixtures/products/product.json"

# 4. Inline fixture references within JSON/XML
data: |
  {
    "user": < ./fixtures/users/user.json,
    "permissions": < ./fixtures/permissions.json
  }
```

**CRITICAL Syntax Rules:**
- `"< file"` - VALID (space after <)
- `"<@ file"` - VALID (@ immediately after <, space after @)
- `"<file"` - INVALID (gracefully falls back to literal string)
- `"< @file"` - INVALID (space between < and @, falls back)

**Security & Performance:**
- Path traversal protection (no `../`)
- Absolute path rejection (no `/etc/passwd` or `C:\`)
- Missing files gracefully fallback to original data
- Thread-safe caching for performance

## TDD/BDD Workflow (MANDATORY)

**Red-Green-Refactor cycle:**
1. **Red**: Write failing test first
2. **Green**: Implement minimum code to pass
3. **Refactor**: Clean up while keeping tests green
4. **Commit**: Run `make check` and commit

**Example:**
```bash
# 1. Write failing test
go test ./pkg/config -v -run TestNewFeature  # ❌ FAIL

# 2. Implement feature in pkg/config/feature.go

# 3. Verify test passes
go test ./pkg/config -v -run TestNewFeature  # ✅ PASS

# 4. Run full quality check
make check  # MUST pass before commit

# 5. Commit (pre-commit hook runs make check automatically)
git add . && git commit -m "feat: add new feature"
```

**NEVER:**
- Commit to `master` or `main` branches (use feature branches)
- Use `//nolint` comments (fix linter issues properly)
- Bypass git hooks (`--no-verify`)

## E2E Test Pattern (MANDATORY)

All E2E tests MUST use **fluent Given/When/Then pattern**.

**Example:**
```go
func TestFixtureFileSupport_AtSyntax_BasicWorkflow(t *testing.T) {
    given, when, then := newParts(t)

    given.
        atSyntaxFixtureConfig()  // Setup

    when.
        a_get_request_is_made_to("/api/users/123")  // Action

    then.
        the_response_is_successful().and().  // Assertions
        the_response_body_contains("John Doe")
}
```

**Key Patterns:**
- **Triple return**: `newParts(t)` returns `given, when, then` (all same `*parts`)
- **Method chaining**: All methods MUST return `*parts` for fluent API
- **Snake_case names**: Descriptive method names (e.g., `a_user_is_created()`)
- **One When, One Then**: Single action per test, then verify
- **`.and()` chaining**: Chain multiple assertions for readability

**DO NOT:**
- Mix Given/When/Then phases
- Use multiple When blocks per test
- Forget to return `*parts` from fluent methods (causes unparam linter errors)

## Key Implementation Details

**Path Matching Priority:**
- Exact matches → Wildcard matches (longer patterns preferred)
- `*` matches one segment: `/users/*` matches `/users/123`
- `**` matches multiple: `/api/**` matches `/api/v1/users/123`

**ID Extraction:**
- **URL paths**: Automatic from wildcard segments
- **JSON bodies**: XPath-like paths (`/id`, `//id`, `/items/*/id`)
- **XML bodies**: Similar XPath syntax
- **Headers**: Multiple supported headers per resource

**Storage Operations:**
- **GET**: Collection (`/users`) or individual (`/users/123`)
- **POST**: Create with auto-generated or extracted IDs
- **PUT**: Update existing (upsert in flexible mode)
- **DELETE**: Remove with conflict detection

**Scenarios:**
- Bypass normal mock behavior
- Used for error testing, edge cases
- Added at runtime via API or in configuration

## Environment Variables

- `UNIMOCK_PORT` - Server port (default: 8080)
- `UNIMOCK_LOG_LEVEL` - Log level (default: info)
- `UNIMOCK_CONFIG` - Config file path (default: config.yaml)

## Testing Strategy

**Unit Tests** (`pkg/`, `internal/`): Fast, isolated component tests
**E2E Tests** (`e2e/`): Full HTTP server lifecycle integration tests
**Configuration Tests**: Validate YAML parsing and behavior
**Performance Tests**: Concurrent request handling

Always test both happy path and error conditions.

## Documentation References

**Essential Reading:**
- `docs/fluent-testing.md` - MANDATORY E2E test pattern (Given/When/Then)
- `docs/pr-guidelines.md` - PR review comment tracking protocol
- `docs/scenarios.md` - Fixture file syntax and scenario configuration
- `docs/configuration.md` - Full configuration reference

**Additional Guides:**
- `docs/testing-guidelines.md` - Unit test patterns, mocking, isolation
- `docs/id_extraction.md` - ID extraction from URLs, bodies, headers
- `docs/client.md` - Go client library usage
- `docs/deployment.md` - Production deployment guides
- `docs/library.md` - Embedding Unimock in applications

**Development Context:**
- `docs/project_structure.md` - Codebase organization
- `docs/decisions.md` - Architectural decision records
- `docs/learnings.md` - Lessons learned from development
