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

[TASK-010] - Done - 2025-05-27 - Implement functionality for scenarios to modify default mock handler status code and response body
[TASK-011] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-001: Scenario overrides default status code and body
[TASK-012] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-002: Scenario overrides default response body, keeps status code
[TASK-013] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-003: Scenario overrides status code, default body used
[TASK-014] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-004: Scenario overrides body, default status code used
[TASK-015] - Skipped - 2025-05-27 - Implement E2E test for SCEN-SMB-005: Default behavior preserved when no relevant scenario is active
[TASK-016] - Skipped - 2025-05-27 - Implement scenario matching logic in mock handler based on RequestPath.
[TASK-017] - Done - 2025-05-27 - Implement E2E test for SCEN-SH-001: Scenario matched by exact RequestPath.
[TASK-018] - Done - 2025-05-27 - Implement E2E test for SCEN-SH-002: Scenario with wildcard in RequestPath is matched.
[TASK-019] - Done - 2025-05-27 - Implement E2E test for SCEN-SH-003: Scenario match skips normal mock handling.
[TASK-020] - ToDo - 2025-05-27 - Implement E2E test for SCEN-SH-004: Scenario for specific HTTP method does not match other methods.
[TASK-021] - ToDo - 2025-05-27 - Implement E2E test for SCEN-SH-005: Scenario matching with empty data and custom location header.
