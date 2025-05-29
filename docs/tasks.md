# Tasks

This file tracks the development tasks for the project.

## Task States
- **ToDo**: The task is pending and has not been started.
- **In Progress**: The task is currently being worked on.
- **Blocked**: The task is blocked by an external factor.
- **Done**: The task has been completed, verified, and all checks have passed.
- **Skipped**: The task has been deemed no longer necessary or will be addressed later.

## Task Format
Each task should follow the format:
`[TASK-ID] - [Status] - [Date Created] - Description`
(e.g., `[TASK-001] - ToDo - 2025-05-27 - Implement user authentication module`)

---

## Current Tasks

[TASK-011] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-001: Scenario overrides default status code and body
[TASK-012] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-002: Scenario overrides default response body, keeps status code
[TASK-013] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-003: Scenario overrides status code, default body used
[TASK-014] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-004: Scenario overrides body, default status code used
[TASK-015] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-005: Default behavior preserved when no relevant scenario is active
[TASK-016] - Skipped - 2025-05-27 - Implement scenario matching logic in mock handler based on RequestPath.
[TASK-030] - Blocked - 2025-05-29 - Implement E2E test for SCEN-E2E-COMPLEX-001: Multistage Resource Lifecycle with Scenario Override. Implementation complete (covered by TASK-029), but blocked by same .hresp tool issue as TASK-029.
[TASK-031] - Done - 2025-05-29 - Fix application bug: Scenario headers defined via POST /_uni/scenarios are not returned when the scenario is matched.

---
ID: TASK-029
Description: Implement a multistage E2E scenario covering resource lifecycle and scenario management.
Status: Blocked - E2E test `TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle` fails due to tool inability to fix whitespace in `e2e/http/scen_e2e_complex_001.hresp`. Manual fix needed for the .hresp file.
Assigned: AI
Created: 2025-05-28
Updated: 2025-05-29
REQ: REQ-E2E-COMPLEX-001 (Implied)
---
