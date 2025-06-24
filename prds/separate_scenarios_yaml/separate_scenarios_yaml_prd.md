# PRD: Separate Scenarios YAML File Support

## Overview
Add optional support for defining scenarios in a separate YAML file instead of only through the runtime API. This will enable pre-configured scenarios for testing and development environments.

## Requirements

### Functional Requirements
- **REQ-001**: Support loading scenarios from a separate YAML file (optional)
- **REQ-002**: Scenarios file should be configurable via environment variable or config
- **REQ-003**: Scenarios from file should be loaded at server startup
- **REQ-004**: Runtime API scenarios should work alongside file-based scenarios
- **REQ-005**: File-based scenarios should follow same format as runtime API
- **REQ-006**: Support for all scenario types (GET, POST, PUT, DELETE, HEAD)

### Technical Requirements
- **TECH-001**: Add scenarios file path to server configuration
- **TECH-002**: Create YAML structure for scenarios definitions
- **TECH-003**: Implement scenarios file loading logic
- **TECH-004**: Integrate file scenarios with existing scenario service
- **TECH-005**: Maintain backward compatibility (optional feature)
- **TECH-006**: Add validation for scenarios file format

## Implementation Plan

### Phase 1: Configuration Updates
1. Add scenarios file path to ServerConfig
2. Add environment variable support for scenarios file
3. Update configuration loading logic

### Phase 2: YAML Structure Design
1. Design scenarios YAML file structure
2. Create example scenarios file
3. Define validation rules

### Phase 3: Loading Implementation
1. Implement scenarios file parsing
2. Add scenarios loading to server startup
3. Integrate with existing scenario service

### Phase 4: Testing
1. Write unit tests for scenarios file loading
2. Write E2E tests with pre-loaded scenarios
3. Test integration with runtime API scenarios

### Phase 5: Validation
1. Run `make check` to ensure all checks pass
2. Manual testing with scenarios file
3. Verify backward compatibility

## YAML Structure Design

```yaml
# scenarios.yaml
scenarios:
  - uuid: "test-scenario-001"
    method: "GET"
    path: "/users/123"
    status_code: 200
    headers:
      Content-Type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }
  
  - uuid: "test-scenario-002"
    method: "POST"
    path: "/users"
    status_code: 201
    headers:
      Content-Type: "application/json"
      Location: "/users/456"
    data: |
      {
        "id": "456",
        "name": "Created User",
        "email": "created@example.com"
      }
```

## Configuration Examples

### Environment Variable
```bash
UNIMOCK_SCENARIOS_FILE=scenarios.yaml
```

### YAML Configuration
```yaml
# config.yaml
server:
  port: 8080
  log_level: info
  scenarios_file: "scenarios.yaml"  # Optional
```

## Acceptance Criteria
- [ ] Scenarios file path configurable via environment variable
- [ ] Scenarios file path configurable via config.yaml
- [ ] Scenarios loaded from file at server startup
- [ ] File scenarios work with all HTTP methods
- [ ] Runtime API scenarios work alongside file scenarios
- [ ] YAML validation prevents invalid scenarios
- [ ] All existing tests pass
- [ ] New tests cover scenarios file functionality
- [ ] `make check` passes successfully
- [ ] Feature is optional (server works without scenarios file)

## Risk Assessment
- **Medium Risk**: Adding new configuration and file loading
- **Mitigation**: Comprehensive testing, optional feature, validation
- **Backward Compatibility**: Maintained (optional feature)