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
Status: ToDo
Assigned: User
Created: 2025-05-30
Updated: 2025-05-30
DependsOn: []
PRDRequirement: All sections in core_unimock_service_prd.md
Notes: Initial review task.
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
Status: Skipped
Assigned: AI (assumed)
Created: 2025-05-27 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md. User to decide if this specific E2E is still needed or covered by other tasks from PRD requirements.
---
ID: CUS-TASK-004 (was TASK-012)
Description: Implement E2E test for SCEN-SMB-002: Scenario overrides default response body, keeps status code.
Status: Skipped
Assigned: AI (assumed)
Created: 2025-05-27 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md. User to decide if this specific E2E is still needed or covered by other tasks from PRD requirements.
---
ID: CUS-TASK-005 (was TASK-013)
Description: Implement E2E test for SCEN-SMB-003: Scenario overrides status code, default body used.
Status: Skipped
Assigned: AI (assumed)
Created: 2025-05-27 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md. User to decide if this specific E2E is still needed or covered by other tasks from PRD requirements.
---
ID: CUS-TASK-006 (was TASK-014)
Description: Implement E2E test for SCEN-SMB-004: Scenario overrides body, default status code used.
Status: Skipped
Assigned: AI (assumed)
Created: 2025-05-27 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md. User to decide if this specific E2E is still needed or covered by other tasks from PRD requirements.
---
ID: CUS-TASK-007 (was TASK-015)
Description: Implement E2E test for SCEN-SMB-005: Default behavior preserved when no relevant scenario is active.
Status: Skipped
Assigned: AI (assumed)
Created: 2025-05-27 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md. User to decide if this specific E2E is still needed or covered by other tasks from PRD requirements.
---
ID: CUS-TASK-008 (was TASK-016)
Description: Implement scenario matching logic in mock handler based on RequestPath.
Status: Skipped
Assigned: AI (assumed)
Created: 2025-05-27 (original)
Updated: 2025-05-30 (migrated)
DependsOn: []
PRDRequirement: Section 4.11 Basic Scenario Handling
Notes: Migrated from old tasks.md. This might be a core implementation task rather than just an E2E test. User to verify if covered by PRD items.
---

*(Further tasks to be derived from the Core Unimock Service PRD requirements.)*
