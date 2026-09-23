# Task Tracker: Core Unimock Service PRD

## PRD Reference

*   [Core Unimock Service PRD](./core_unimock_service_prd.md)

## Task States

*   **ToDo**: The task is pending and has not been started.
*   **In Progress**: The task is currently being worked on.
*   **Blocked**: The task is blocked by an external factor.
*   **Done**: The task has been completed, verified, and all checks (including tests) have passed.
*   **Skipped**: The task has been deemed no longer necessary or will be addressed later.

## Task Format

Each task should follow the format:

```
---
ID: [PRD_PREFIX-TASK-XXX] (e.g., CUS-TASK-001)
Description: Brief description of the task.
Status: ToDo | In Progress | Blocked | Done | Skipped
Assigned: [Assignee Name/Team/AI]
Created: YYYY-MM-DD
Updated: YYYY-MM-DD
DependsOn: [List of Task IDs this task depends on, if any]
PRDRequirement: [Link or reference to specific requirement(s) in the PRD]
Notes: Optional notes or comments.
---
```

## Tasks

---
ID: CUS-TASK-001
Description: Review and refine all requirements in the Core Unimock Service PRD for clarity, consistency, and testability.
Status: Done
Assigned: AI
Created: 2025-05-30
Updated: 2025-06-23
DependsOn: []
PRDRequirement: All sections in core_unimock_service_prd.md
Notes: Completed comprehensive review and refinement. Key improvements: clarified multi-ID support, resolved collection response formatting inconsistencies, added ID extraction priority order, updated out-of-scope section, enhanced security and configuration specifications.
---
ID: CUS-TASK-002 (was TASK-031)
Description: Fix application bug: Scenario headers defined via POST /_uni/scenarios are not returned when the scenario is matched.
Status: Done
Assigned: AI (assumed)
Created: 2025-05-29 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md.
---
ID: CUS-TASK-003 (was TASK-011)
Description: Implement E2E test for SCEN-SMB-001: Scenario overrides default status code and body.
Status: ToDo
Assigned: AI
Created: 2025-05-27 (original)
Updated: 2025-06-23 (reviewed)
DependsOn: [CUS-TASK-009]
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: CRITICAL: Analysis shows scenario-mock integration is missing from mock handler. This test covers unique functionality not covered by existing tests. Should be implemented after core scenario checking logic is added to mock handler.
---
ID: CUS-TASK-004 (was TASK-012)
Description: Implement E2E test for SCEN-SMB-002: Scenario overrides default response body, keeps status code.
Status: ToDo
Assigned: AI
Created: 2025-05-27 (original)
Updated: 2025-06-23 (reviewed)
DependsOn: [CUS-TASK-009]
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Should be implemented after core scenario checking logic is added to mock handler. Tests partial scenario override functionality.
---
ID: CUS-TASK-005 (was TASK-013)
Description: Implement E2E test for SCEN-SMB-003: Scenario overrides status code, default body used.
Status: ToDo
Assigned: AI
Created: 2025-05-27 (original)
Updated: 2025-06-23 (reviewed)
DependsOn: [CUS-TASK-009]
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Should be implemented after core scenario checking logic is added to mock handler. Tests partial scenario override functionality.
---
ID: CUS-TASK-006 (was TASK-014)
Description: Implement E2E test for SCEN-SMB-004: Scenario overrides body, default status code used.
Status: ToDo
Assigned: AI
Created: 2025-05-27 (original)
Updated: 2025-06-23 (reviewed)
DependsOn: [CUS-TASK-009]
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Should be implemented after core scenario checking logic is added to mock handler. Tests partial scenario override functionality.
---
ID: CUS-TASK-007 (was TASK-015)
Description: Implement E2E test for SCEN-SMB-005: Default behavior preserved when no relevant scenario is active.
Status: ToDo
Assigned: AI
Created: 2025-05-27 (original)
Updated: 2025-06-23 (reviewed)
DependsOn: [CUS-TASK-009]
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Should be implemented after core scenario checking logic is added to mock handler. Tests default behavior when no scenario matches.
---
ID: CUS-TASK-008 (was TASK-016)
Description: Implement scenario matching logic in mock handler based on RequestPath.
Status: ToDo
Assigned: AI
Created: 2025-05-27 (original)
Updated: 2025-06-23 (reviewed)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: CRITICAL: Core implementation missing. Mock handler does not currently check for scenarios. This is a prerequisite for SCEN-SMB tests.
---

---
ID: CUS-TASK-009
Description: Implement scenario checking logic in mock handler to support scenario-mock integration.
Status: ToDo
Assigned: AI
Created: 2025-06-23
Updated: 2025-06-23
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: CRITICAL missing functionality. Mock handler must check for scenarios by RequestPath before processing normal mock logic. This is a prerequisite for all SCEN-SMB tests.
---

*(Further tasks to be derived from the Core Unimock Service PRD requirements.)*
