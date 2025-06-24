# PRD: Configuration Flag for POST/UPDATE/DELETE Response Body Control

## Overview
Add a configuration flag to control whether POST/UPDATE/DELETE operations return the resource body in the response or return an empty body.

## Requirements

### Functional Requirements
- **REQ-001**: Add `return_body` boolean flag to configuration structure
- **REQ-002**: Default value for `return_body` flag must be `false` 
- **REQ-003**: When `return_body` is `true`, POST/UPDATE/DELETE operations return the resource body
- **REQ-004**: When `return_body` is `false`, POST/UPDATE/DELETE operations return empty body
- **REQ-005**: Configuration flag must be configurable via YAML configuration file

### Technical Requirements
- **TECH-001**: Modify existing configuration structures to include the flag
- **TECH-002**: Update handler layer to check flag before returning responses
- **TECH-003**: Maintain backward compatibility with existing configurations
- **TECH-004**: Add comprehensive test coverage for both flag states

## Implementation Plan

### Phase 1: Configuration Structure
1. Add `return_body` field to appropriate config struct
2. Ensure default value is `false`
3. Update YAML parsing to include new field

### Phase 2: Handler Updates
1. Modify POST handler to conditionally return body
2. Modify UPDATE handler to conditionally return body  
3. Modify DELETE handler to conditionally return body
4. Pass configuration to handlers as needed

### Phase 3: Testing
1. Write unit tests for configuration parsing
2. Write handler tests for both flag states
3. Write E2E tests to verify end-to-end behavior

### Phase 4: Validation
1. Run `make check` to ensure all checks pass
2. Manual testing with sample configurations

## Acceptance Criteria
- [x] Configuration flag `return_body` exists with default `false`
- [x] POST operations respect flag setting
- [x] UPDATE operations respect flag setting
- [x] DELETE operations respect flag setting
- [x] All existing tests pass
- [x] New tests cover both flag states
- [x] `make check` passes successfully
- [x] Backward compatibility maintained

## Risk Assessment
- **Low Risk**: Simple boolean flag addition
- **Mitigation**: Comprehensive testing and backward compatibility checks