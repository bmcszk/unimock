# Fixture File Support PRD

## Problem
Users need to reference external fixture files in scenario configurations instead of inlining response data. This enables better separation of configuration and test data, making configs cleaner and more maintainable.

## Solution
Add support for `@fixtures/` file references in scenario configuration that loads response body content from external files.

## Requirements

### Functional Requirements
1. **File Reference Syntax**: Support `@fixtures/path/to/file.ext` syntax in scenario `data` field
2. **File Types**: Support JSON, XML, and text fixture files
3. **Base Path**: Resolve fixture paths relative to configuration file directory
4. **Fallback**: Maintain backward compatibility with inline data
5. **Error Handling**: Return clear error messages for missing/invalid fixture files

### Non-Functional Requirements
1. **Performance**: Cache loaded fixture files
2. **Security**: Restrict file access to fixture directories
3. **Backwards Compatibility**: Existing configs continue working unchanged

## Implementation Plan

### Phase 1: Core Functionality
1. Create fixture file resolver service
2. Add file detection logic (detects `@` prefix)
3. Implement file loading with caching
4. Add error handling for invalid files

### Phase 2: Integration
1. Update `ScenarioConfig.ToModelScenario()` to resolve file references
2. Update configuration loading to pass base directory
3. Add tests for fixture resolution

### Phase 3: Edge Cases
1. Add security validation (path traversal protection)
2. Handle missing files gracefully
3. Add support for nested directory structures

## Task List

### Development Tasks
- [ ] Create `pkg/config/fixture_resolver.go`
- [ ] Add tests in `pkg/config/fixture_resolver_test.go`
- [ ] Update `ScenarioConfig.ToModelScenario()` method
- [ ] Update config loading to track base directory
- [ ] Add integration tests
- [ ] Add documentation examples

### Testing Tasks
- [ ] Unit tests for fixture resolver
- [ ] Integration tests with scenario loading
- [ ] Error handling tests
- [ ] Backward compatibility tests
- [ ] Performance tests for caching

## Acceptance Criteria

1. ✅ Configuration with `data: "@fixtures/robots.json"` loads file content
2. ✅ File paths resolved relative to config file location
3. ✅ Missing files return clear error messages
4. ✅ Inline data continues working unchanged
5. ✅ Security prevents path traversal attacks
6. ✅ All existing tests pass
7. ✅ New tests achieve >90% coverage

## Technical Details

### Configuration Example
```yaml
sections:
  operations:
    path_pattern: "/internal/robots*"
    return_body: true

scenarios:
  - uuid: "list-robots"
    method: "GET"
    path: "/internal/robots"
    status_code: 200
    data: "@fixtures/operations/robots.json"
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