# Fixture File Support PRD

## Problem
Users need to reference external fixture files in scenario configurations instead of inlining response data. This enables better separation of configuration and test data, making configs cleaner and more maintainable.

## Solution
Add comprehensive fixture file support for scenario configurations that supports:
- `@fixtures/` syntax (backward compatibility)
- `< ./fixtures/file.ext` syntax (go-restclient compatible)
- Inline fixture references within body content using `< ./fixtures/file.ext` syntax
- Thread-safe caching for performance
- Security validation for path traversal protection

## Requirements

### Functional Requirements
1. **File Reference Syntax**:
   - `@fixtures/path/to/file.ext` syntax (backward compatibility)
   - `< ./fixtures/path/to/file.ext` syntax (go-restclient compatible)
   - `<@ ./fixtures/path/to/file.ext` syntax (future variable substitution)
   - Inline references: `{"user": < ./fixtures/user.json, "status": "active"}`
2. **File Types**: Support JSON, XML, and text fixture files
3. **Base Path**: Resolve fixture paths relative to configuration file directory
4. **Fallback**: Maintain backward compatibility with inline data
5. **Error Handling**: Return clear error messages for missing/invalid fixture files
6. **Inline Support**: Allow mixing inline data with fixture references in body content

### Non-Functional Requirements
1. **Performance**: Cache loaded fixture files
2. **Security**: Restrict file access to fixture directories
3. **Backwards Compatibility**: Existing configs continue working unchanged

## Implementation Plan (TDD Approach)

### Phase 1: Core Fixture Resolver (TDD)
1. Create initial failing tests for basic `@` syntax
2. Implement minimal fixture resolver to pass tests
3. Add tests for `<` syntax support
4. Extend resolver to handle both syntaxes
5. Add tests for file caching and thread safety
6. Implement caching with mutex protection

### Phase 2: Inline Fixture Support (TDD)
1. Create failing tests for inline fixture detection
2. Implement regex-based fixture detection in body content
3. Add tests for multiple inline fixtures in single response
4. Implement inline fixture resolution
5. Add tests for mixed inline/fixture content
6. Implement content replacement logic

### Phase 3: Security and Error Handling (TDD)
1. Create failing tests for path traversal attacks
2. Implement security validation
3. Add tests for missing file handling
4. Implement graceful error handling
5. Add tests for invalid file paths
6. Implement comprehensive validation

### Phase 4: Integration (TDD)
1. Create failing tests for scenario integration
2. Update `ScenarioConfig.ToModelScenario()` to use resolver
3. Add tests for configuration loading with base directory
4. Implement configuration integration
5. Create E2E tests for complete workflow
6. Verify backward compatibility

## Task List (TDD-Based)

### Phase 1: Core Fixture Resolver
- [ ] **TDD-01**: Create failing test for `@fixtures/file.json` syntax
- [ ] **TDD-02**: Implement basic resolver to pass `@` syntax test
- [ ] **TDD-03**: Create failing test for `< ./fixtures/file.json` syntax
- [ ] **TDD-04**: Extend resolver to support `<` syntax
- [ ] **TDD-05**: Create failing test for `<@ ./fixtures/file.json` syntax
- [ ] **TDD-06**: Extend resolver to support `<@` syntax
- [ ] **TDD-07**: Create failing test for file caching
- [ ] **TDD-08**: Implement thread-safe caching with mutex
- [ ] **TDD-09**: Create failing test for missing files
- [ ] **TDD-10**: Implement graceful missing file handling

### Phase 2: Inline Fixture Support
- [ ] **TDD-11**: Create failing test for inline `{"key": < ./file.json}` syntax
- [ ] **TDD-12**: Implement regex-based inline detection
- [ ] **TDD-13**: Create failing test for multiple inline fixtures
- [ ] **TDD-14**: Implement multiple inline fixture resolution
- [ ] **TDD-15**: Create failing test for mixed content
- [ ] **TDD-16**: Implement mixed inline/regular content handling

### Phase 3: Security and Validation
- [ ] **TDD-17**: Create failing test for path traversal `@../../../etc/passwd`
- [ ] **TDD-18**: Implement path traversal protection
- [ ] **TDD-19**: Create failing test for absolute paths `@/etc/passwd`
- [ ] **TDD-20**: Implement absolute path blocking
- [ ] **TDD-21**: Create failing test for empty references `@`
- [ ] **TDD-22**: Implement empty reference validation

### Phase 4: Integration and E2E
- [ ] **TDD-23**: Create failing test for scenario config integration
- [ ] **TDD-24**: Update `ScenarioConfig.ToModelScenario()` to use resolver
- [ ] **TDD-25**: Create failing E2E test for complete workflow
- [ ] **TDD-26**: Implement server integration with config loading
- [ ] **TDD-27**: Create E2E test for complex nested inline fixtures
- [ ] **TDD-28**: Verify backward compatibility with existing tests

## Acceptance Criteria

1. [ ] Configuration with `data: "@fixtures/robots.json"` loads file content
2. [ ] Configuration with `data: "< ./fixtures/robots.json"` loads file content
3. [ ] Configuration with `data: "<@ ./fixtures/robots.json"` loads file content
4. [ ] Configuration with inline `{"user": < ./fixtures/user.json}` resolves fixtures
5. [ ] Configuration with multiple inline fixtures works correctly
6. [ ] File paths resolved relative to config file location
7. [ ] Missing files return clear error messages or preserve inline references
8. [ ] Inline data continues working unchanged
9. [ ] Security prevents path traversal attacks for all syntax types
10. [ ] All existing tests pass
11. [ ] New tests achieve >90% coverage
12. [ ] E2E tests verify complete fixture file workflow including JSON, XML, and text files
13. [ ] Caching improves performance for repeated fixture access
14. [ ] Thread-safe operation under concurrent access

## Technical Details

### Configuration Examples

#### Basic Fixture References
```yaml
sections:
  operations:
    path_pattern: "/internal/robots*"
    return_body: true

scenarios:
  # Backward compatible @ syntax
  - uuid: "list-robots-at"
    method: "GET"
    path: "/internal/robots"
    status_code: 200
    data: "@fixtures/operations/robots.json"

  # New go-restclient compatible < syntax
  - uuid: "list-robots-less"
    method: "GET"
    path: "/internal/robots"
    status_code: 200
    data: "< ./fixtures/operations/robots.json"

  # Future variable substitution <@ syntax
  - uuid: "list-robots-less-at"
    method: "GET"
    path: "/internal/robots"
    status_code: 200
    data: "<@ ./fixtures/operations/robots.json"
```

#### Inline Fixture References
```yaml
scenarios:
  # Single inline fixture
  - uuid: "user-profile"
    method: "GET"
    path: "/api/users/123"
    data: |
      {
        "user": < ./fixtures/users/user_123.json,
        "metadata": {
          "timestamp": "2024-01-15T10:30:00Z",
          "endpoint": "user-profile"
        }
      }

  # Multiple inline fixtures
  - uuid: "complete-user-data"
    method: "GET"
    path: "/api/users/123/complete"
    data: |
      {
        "user": < ./fixtures/users/user_123.json,
        "profile": < ./fixtures/users/user_123_profile.json,
        "permissions": < ./fixtures/permissions/user_123.json,
        "settings": {
          "theme": "dark",
          "language": "en"
        }
      }

  # Complex nested with inline fixtures
  - uuid: "complex-response"
    method: "POST"
    path: "/api/reports/generate"
    data: |
      {
        "response": {
          "user_data": < ./fixtures/users/user_123.json,
          "report_data": < ./fixtures/reports/monthly_summary.json,
          "system_info": {
            "generated_at": "2024-01-15T10:30:00Z",
            "version": "v2.1"
          }
        }
      }
```

### File Structure
```
config.yaml
fixtures/
  operations/
    robots.json
    status_C10190.json
```

### API Changes
- `pkg/config/uni_config.go`: Add fixture resolver integration
- `pkg/config/fixture_resolver.go`: New file resolution service
- `ScenarioConfig.ToModelScenario()`: Enhanced to resolve file references

## Definition of Done
- All acceptance criteria met
- Tests passing with >90% coverage
- Documentation updated with examples
- No breaking changes to existing functionality
