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

[TASK-022] - Done - 2025-05-28 - Design: Data structure for storing resources with multiple IDs.
[TASK-023] - Done - 2025-05-28 - Implement: Update storage layer (mock_storage.go) to support multiple IDs per resource.
[TASK-024] - Done - 2025-05-28 - Implement: Update ID extraction logic in mock_handler.go.
[TASK-025] - Skipped - 2025-05-28 - Config: Add configuration options for specifying multiple ID headers and body paths.
[TASK-026] - Done - 2025-05-28 - Test: Write unit tests for multiple ID functionality in storage and handler.
[TASK-027] - Done - 2025-05-28 - Scenario: Define test scenarios for REQ-RM-MULTI-ID in docs/scenarios.md.
[TASK-028] - Done - 2025-05-28 - Test: Implement E2E tests for REQ-RM-MULTI-ID.

