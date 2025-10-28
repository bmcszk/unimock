# Comprehensive E2E Test List

This document catalogues all End-to-End (E2E) tests in the unimock project and tracks their migration status to the fluent pattern.

## Migration Status Legend

- ✅ **Already Fluent** - Test uses given/when/then pattern with `newParts()`
- 🔄 **Needs Migration** - Test needs to be converted to fluent pattern
- 📝 **Files Using Fluent Pattern** - `e2e_parts_test.go`, `e2e_server_parts_test.go`

## Test Files Overview

### Already Fluent (✅)
1. `e2e_test.go` - Multiple ID resource management tests
2. `complex_e2e_scenario_test.go` - Complex lifecycle tests with scenarios
3. `request_handling_test.go` - Request handling and scenario tests
4. `universal_client_test.go` - Universal client E2E tests

### Need Migration (🔄)
1. `scenarios_file_test.go` - Scenario file loading tests
2. `scenarios_only_test.go` - Scenarios-only configuration tests
3. `complex_lifecycle_test.go` - Complex lifecycle test

---

## Detailed Test Inventory

### `e2e/e2e_test.go` ✅ (Already Fluent)
**Pattern**: Uses `newParts()` and fluent chaining
**Test Count**: 5 tests

| Test Function | Description | Given | When | Then |
|---------------|-------------|-------|------|------|
| `TestE2E_SCEN_RM_MULTI_ID_001` | Create resource with multiple IDs | - | `the_resource_is_created_with_multiple_ids()` | `the_resource_can_be_retrieved_by_either_id()` |
| `TestE2E_SCEN_RM_MULTI_ID_002` | Update resource with verification | - | `the_resource_is_updated_and_verified()` | `the_update_is_successful()` |
| `TestE2E_SCEN_RM_MULTI_ID_003` | Delete resource with verification | - | `the_resource_is_deleted_and_verified()` | `the_deletion_is_successful()` |
| `TestE2E_SCEN_RM_MULTI_ID_004` | Conflict error for conflicting IDs | - | `a_resource_is_created_with_conflicting_ids()` | `a_conflict_error_is_returned()` |
| `TestE2E_SCEN_RM_MULTI_ID_005` | Multiple ID retrieval (alternative) | - | `the_resource_is_created_with_multiple_ids()` | `the_resource_can_be_retrieved_by_either_id()` |

### `e2e/complex_e2e_scenario_test.go` ✅ (Already Fluent)
**Pattern**: Uses `newParts()` with scenario overrides
**Test Count**: 1 test with 3 parts

| Test Function | Description | Given | When | Then |
|---------------|-------------|-------|------|------|
| `TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle` | Complex resource lifecycle with scenario override | `an_http_request_is_made_from_file()` + validation | `a_scenario_override_is_applied()` + request execution | `the_scenario_override_is_deleted()` + final validation |

### `e2e/request_handling_test.go` ✅ (Already Fluent)
**Pattern**: Uses `newParts()` with scenario setup
**Test Count**: 22 tests

**Request Handling Tests (SCEN-RH-001 to SCEN-RH-010)**:
- `TestSCEN_RH_001_GetExistingResource` - Get existing resource
- `TestSCEN_RH_002_PostCreateResource` - Create new resource via POST
- `TestSCEN_RH_003_PutUpdateResource` - Update resource via PUT
- `TestSCEN_RH_004_DeleteResource` - Delete resource via DELETE
- `TestSCEN_RH_005_GetIndividualResourceEndpoint` - Individual resource endpoint
- `TestSCEN_RH_006_GetCollectionEndpoint` - Collection endpoint
- `TestSCEN_RH_007_PostInvalidContentType` - Invalid content type rejection
- `TestSCEN_RH_008_GetNonExistentResource` - Non-existent resource returns 404
- `TestSCEN_RH_009_PathBasedRouting` - Path-based routing
- `TestSCEN_RH_010_WildcardPathMatching` - Wildcard path matching

**Scenario Handling Tests (SCEN-SH-001 to SCEN-SH-005)**:
- `TestSCEN_SH_001_ExactPathScenarioMatch` - Exact path scenario matching
- `TestSCEN_SH_002_WildcardPathScenarioMatch` - Wildcard path scenario matching
- `TestSCEN_SH_003_ScenarioSkipsMockHandling` - Scenario skips mock handling
- `TestSCEN_SH_004_ScenarioMethodMismatch` - Method-specific scenarios
- `TestSCEN_SH_005_ScenarioWithEmptyDataAndLocation` - Empty data with location header

### `e2e/universal_client_test.go` ✅ (Already Fluent)
**Pattern**: Uses `newParts()` with comprehensive client operations
**Test Count**: 1 test with 3 subtests

- `TestUniversalClientE2E` - `BasicHTTPOperations` - Basic HTTP operations (GET, HEAD, OPTIONS)
- `TestUniversalClientE2E` - `JSONOperations` - JSON CRUD operations
- `TestUniversalClientE2E` - `UniDataLifecycle` - Collection lifecycle

### `e2e/scenarios_file_test.go` 🔄 (Needs Migration)
**Current Pattern**: Uses `newServerParts()` but not strict given/when/then
**Test Count**: 2 tests with multiple subtests

| Test Function | Description | Current Structure | Target Structure |
|---------------|-------------|-------------------|------------------|
| `TestScenarioFileLoading` | Load scenarios from unified config | Uses `newServerParts()` with subtests | Convert to strict given/when/then |
| `TestScenarioFileAndRuntimeAPIIntegration` | File + runtime scenario integration | Uses `newServerParts()` with helper methods | Convert to strict given/when/then |

**Subtests to migrate**:
- `GET scenario from file`
- `POST scenario from file`
- `HEAD scenario from file`
- `File scenario works`
- `Runtime API scenario works alongside file scenario`

### `e2e/scenarios_only_test.go` 🔄 (Needs Migration)
**Current Pattern**: Traditional test structure without fluent pattern
**Test Count**: 1 test with 5 subtests

| Test Function | Description | Current Structure | Target Structure |
|---------------|-------------|-------------------|------------------|
| `TestScenariosOnlyConfiguration` | Scenarios-only configuration | Traditional setup + subtests | Convert to given/when/then with `newParts()` |

**Subtests to migrate**:
- `GET scenario returns expected response`
- `POST scenario returns expected response with location header`
- `Error scenario returns expected error response`
- `Non-scenario path returns 404 (no sections configured)`
- `Health endpoint still works`

### `e2e/complex_lifecycle_test.go` 🔄 (Needs Migration)
**Current Pattern**: Minimal fluent pattern (no given part)
**Test Count**: 1 test

| Test Function | Description | Current Structure | Target Structure |
|---------------|-------------|-------------------|------------------|
| `TestComplexLifecycle` | Complex lifecycle test | Uses `_, when, then := newParts(t)` | Add proper given part if needed |

---

## Migration Summary

**Total Tests**: 32 tests across 7 files
- **Already Fluent**: 27 tests (84%) across 4 files
- **Need Migration**: 5 tests (16%) across 3 files

**Files requiring migration**:
1. `scenarios_file_test.go` - Convert server-based tests to strict fluent pattern
2. `scenarios_only_test.go` - Complete rewrite to use `newParts()` and fluent chaining
3. `complex_lifecycle_test.go` - Enhance with proper given section

**Migration Requirements**:
- All tests must have exactly one //given //when //then section
- Use existing fluent helper methods where possible
- Leverage `go-restclient` integration for HTTP operations
- Preserve existing test logic and assertions