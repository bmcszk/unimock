# PRD: HEAD Operation Support

## Overview
Add HEAD HTTP method support to Unimock. HEAD should behave exactly like GET but return no response body.

## Requirements

### Functional Requirements
- **REQ-001**: Support HEAD HTTP method for all configured path patterns
- **REQ-002**: HEAD requests should return same headers as equivalent GET requests
- **REQ-003**: HEAD requests should return same status codes as equivalent GET requests  
- **REQ-004**: HEAD requests must return empty response body
- **REQ-005**: HEAD should work with all existing features (scenarios, strict paths, etc.)

### Technical Requirements
- **TECH-001**: Update router to handle HEAD method
- **TECH-002**: Modify mock handler to process HEAD requests
- **TECH-003**: Ensure HEAD works with scenario system
- **TECH-004**: Add comprehensive test coverage for HEAD method
- **TECH-005**: Maintain backward compatibility

## Implementation Plan

### Phase 1: Router Updates
1. Add HEAD method to router method handling
2. Route HEAD requests to mock handler

### Phase 2: Handler Implementation  
1. Modify mock handler to support HEAD method
2. Implement HEAD logic (same as GET but no body)
3. Ensure scenario compatibility

### Phase 3: Testing
1. Write unit tests for HEAD operation
2. Write E2E tests for HEAD behavior
3. Test with scenarios and strict paths

### Phase 4: Validation
1. Run `make check` to ensure all checks pass
2. Manual testing with HEAD requests

## Acceptance Criteria
- [x] HEAD method accepted by router
- [x] HEAD requests processed by mock handler
- [x] HEAD returns same headers/status as GET
- [x] HEAD returns empty body
- [x] HEAD works with scenarios
- [x] HEAD works with strict paths
- [x] All existing tests pass
- [x] New tests cover HEAD behavior
- [x] `make check` passes successfully

## Risk Assessment
- **Low Risk**: Simple addition following existing GET pattern
- **Mitigation**: Comprehensive testing and following HTTP standards