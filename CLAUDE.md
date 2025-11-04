# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

**You have two weapons, use them often!**
- `make check` is your primary weapon, it should be quick
- `make test-all` is your secondary weapon, it should run all tests including e2e

**Primary Quality Gate:**
```bash
make check    # MUST run before commits - runs vet, lint, unit tests
```

**Core Development:**
```bash
make build     # Build the unimock binary
make run       # Build and run locally
make test-all  # Run all tests (unit + E2E)
make test-unit # Unit tests with race detection
make test-e2e  # End-to-end tests with actual HTTP server
```

**Code Quality:**
```bash
make vet       # Go vet static analysis
make lint      # golangci-lint with strict rules
make tidy      # Clean up dependencies
```

**Testing Individual Components:**
```bash
go test ./pkg/config -v -run TestSpecificFunction  # Run specific test
go test ./internal/service -v                   # Test specific package
```

**Single Test with Verbose Output:**
```bash
gotestsum -- ./pkg/config -v -run TestUniConfig_LoadFromYAML
```

## Architecture Overview

Unimock is a **universal HTTP mock server** that implements a clean layered architecture:

```
HTTP Request → Chi Router → Handler → Service → Storage → Response
```

### Core Components

**Configuration System (`pkg/config/`)**:
- **Unified YAML format** supporting both `sections` (mock behavior) and `scenarios` (predefined responses)
- **Legacy format compatibility** for backward compatibility
- **Smart ID extraction** from URLs, JSON/XML bodies, and HTTP headers
- **Wildcard pattern matching** with `*` (single segment) and `**` (recursive) wildcards
- **Fixture file support** with `@fixtures/` syntax for external files

**Request Flow**:
1. **Router** (`internal/router/`) - Chi v5 with middleware and scenario override logic
2. **Handler** (`internal/handler/`) - UniHandler for CRUD operations, ScenarioHandler for overrides, TechHandler for health/metrics
3. **Service** (`internal/service/`) - Business logic with conflict detection and scenario matching
4. **Storage** (`internal/storage/`) - Thread-safe in-memory storage with composite keys

**Key Storage Concepts**:
- **Composite key system**: Resources identified by multiple IDs (URL path, body content, headers)
- **Strict vs Flexible path modes**: Enforce creation path structure or allow cross-path access
- **Conflict detection**: Prevent simultaneous modifications to the same resource

## Configuration Patterns

**Section Configuration** (defines mock behavior):
```yaml
sections:
  users:
    path_pattern: "/api/users/*"           # Wildcard matching
    body_id_paths: ["/id", "/user/id"]    # JSON/XML extraction
    header_id_names: ["X-User-ID"]      # Header extraction
    return_body: true                      # Response body control
    strict_path: false                    # Path access mode
```

**Scenario Overrides** (predefined responses):
```yaml
scenarios:
  - uuid: "user-not-found"
    method: "GET"
    path: "/api/users/999"
    status_code: 404
    data: '{"error": "User not found"}'
```

**Fixture File Support** - Three syntax styles (go-restclient compatible):
```yaml
# 1. Legacy @ syntax (backward compatibility)
data: "@fixtures/operations/robots.json"

# 2. Space-after-< syntax (SPACE REQUIRED after <)
data: "< ./fixtures/users/user.json"

# 3. <@ syntax (@ IMMEDIATELY after <, no space between)
data: "<@ ./fixtures/products/product.json"

# 4. Inline fixture references within JSON/XML
data: |
  {
    "user": < ./fixtures/users/user.json,
    "permissions": < ./fixtures/permissions.json
  }

# IMPORTANT: Invalid syntax gracefully falls back to literal string
# - "<./file" (no space) → returned as-is
# - "< @./file" (space between < and @) → returned as-is
```

**Fixture Resolver Security**:
- Path traversal protection (no `../` allowed)
- Absolute path rejection (no `/etc/passwd` or `C:\`)
- File-not-found errors gracefully fallback to original data
- Caching for performance (thread-safe)

## Development Workflow

1. **Always run `make check` before commits** - it's the primary quality gate
2. **Git hooks enforce quality**:
   - Pre-commit: Runs `make check` (vet, lint, unit tests)
   - Pre-push: Runs `make test-all` (unit + E2E tests)
3. **E2E tests start a real server** and make actual HTTP requests
4. **Configuration-first approach** - behavior defined in YAML, not code
5. **Test with both unit and E2E levels** - unit for logic, E2E for integration

## E2E Test Pattern (MANDATORY)

All E2E tests MUST use the **fluent Given/When/Then pattern** from `docs/fluent-testing.md`:

```go
func TestFixtureFileSupport_AtSyntax_BasicWorkflow(t *testing.T) {
    given, when, then := newParts(t)

    given.
        atSyntaxFixtureConfig()  // Setup config with fixtures

    when.
        a_get_request_is_made_to("/api/users/123")

    then.
        the_response_is_successful().and().
        the_response_body_contains("John Doe")
}
```

**Key patterns**:
- **Triple return**: `newParts(t)` returns `given, when, then` (all point to same `*parts`)
- **Method chaining**: All methods return `*parts` for fluent API
- **Snake_case names**: Test methods use descriptive snake_case (e.g., `a_user_is_created()`)
- **One When, One Then**: Each test should have exactly one action (when) and verify it (then)
- **`.and()` for clarity**: Chain multiple assertions with `.and()` for readability

**DO NOT**:
- Mix Given/When/Then phases
- Use multiple When blocks in one test
- Forget to return `*parts` from fluent methods (causes linter errors)

## Key Implementation Details

**Path Matching Priority**: Exact matches → Wildcard matches (longer patterns preferred)
- `*` matches one segment: `/users/*` matches `/users/123`
- `**` matches multiple segments: `/api/**` matches `/api/v1/users/123`

**ID Extraction Methods**:
- **URL paths**: Automatic from wildcard segments
- **JSON bodies**: XPath-like paths (`/id`, `//id`, `/items/*/id`)
- **XML bodies**: Similar XPath syntax
- **Headers**: Multiple supported headers per resource

**Storage Operations**:
- **GET**: Collection access (`/users`) or individual resource (`/users/123`)
- **POST**: Create new resources with auto-generated or extracted IDs
- **PUT**: Update existing resources (upsert in flexible mode)
- **DELETE**: Remove resources with conflict detection

**Scenario System**:
- Scenarios bypass normal mock behavior
- Used for error testing, edge cases, and specific test conditions
- Can be added at runtime via API or defined in configuration

## Environment Variables

- `UNIMOCK_PORT` - Server port (default: 8080)
- `UNIMOCK_LOG_LEVEL` - Log level (default: info)
- `UNIMOCK_CONFIG` - Config file path (default: config.yaml)

## Testing Strategy

**Unit Tests**: Fast tests for individual components and logic (in `pkg/` and `internal/`)
**E2E Tests**: Integration tests with full HTTP server lifecycle (in `e2e/`)
**Configuration Tests**: Validate YAML parsing and behavior
**Performance Tests**: Concurrent request handling and storage operations

Always test both the happy path and error conditions for new features.

## TDD/BDD Workflow (REQUIRED)

Follow **Test-Driven Development** for all new features:

1. **Red**: Write failing test first
2. **Green**: Implement minimum code to pass
3. **Refactor**: Clean up while keeping tests green
4. **Commit**: Run `make check` and commit

**Example TDD cycle**:
```bash
# 1. Write failing test
go test ./pkg/config -v -run TestNewFeature  # ❌ FAIL

# 2. Implement feature
# Edit pkg/config/feature.go

# 3. Verify test passes
go test ./pkg/config -v -run TestNewFeature  # ✅ PASS

# 4. Run full quality check
make check  # Must pass before commit

# 5. Commit (pre-commit hook runs make check automatically)
git add . && git commit -m "feat: add new feature"
```

**Important**:
- NEVER commit to `master` or `main` branches (use feature branches)
- NEVER use `//nolint` comments (fix linter issues properly)
- NEVER bypass git hooks (`--no-verify`)

