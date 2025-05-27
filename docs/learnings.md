# Learnings Log

This document records mistakes made by the AI assistant during development, along with their resolutions. This serves as a continuous improvement log.

---

## Log Entries

**Date:** 2025-05-27
**Mistake:** The `edit_file` tool (and `reapply`) repeatedly failed to correctly apply changes to `docs/tasks.md` when instructed to replace a section or the entire content. The tool's response indicated no change or an incorrect partial change, even though the user-provided file state later showed the changes were eventually successful. This suggests a potential discrepancy between the tool's execution/feedback and the file system state, or a difficulty in handling full-file replacement edits for this specific file or pattern of edits.
**File(s) Affected:** `docs/tasks.md`
**Resolution:** The issue was implicitly resolved as the file system eventually reflected the correct state, possibly due to user intervention or a delayed effect of the tool. The learning is to be cautious with `edit_file` for full content replacement and to verify the actual file state if the tool reports issues, and to consider alternative edit strategies if problems persist. For this instance, proceeding as if the file is correct based on user-provided context.

**Date:** 2025-05-27
**Mistake:** Incorrectly placed the `Scenario` model definition in `internal/model/scenario.go` instead of `pkg/model/scenario.go`. This led to import conflicts and linter errors in `internal/handler/mock_handler.go`.
**File(s) Affected:** `internal/model/scenario.go`, `pkg/model/scenario.go`, `internal/handler/mock_handler.go`
**Resolution:** Moved the `Scenario`, `ScenarioRequest`, and `ScenarioResponse` struct definitions to `pkg/model/scenario.go` and updated imports in `internal/handler/mock_handler.go` accordingly. Deleted the incorrect `internal/model/scenario.go`.

**Date:** 2025-05-27
**Mistake:** Incorrectly refactored the `pkg/model.Scenario` struct into sub-structs (`ScenarioRequest`, `ScenarioResponse`). The user clarified that the original flatter structure (with fields like `UUID`, `RequestPath`, `StatusCode`, `Body`, `ContentType`, `Location` directly on `Scenario`) was correct and intended for both request matching and defining the response override. This also caused a mismatch with `internal/service/scenario_service.go`.
**File(s) Affected:** `pkg/model/scenario.go`, `internal/handler/mock_handler.go`
**Resolution:** Restored `pkg/model/scenario.go` to the correct flatter structure. Updated `internal/handler/mock_handler.go` to parse `RequestPath` and use the direct fields from `Scenario` for applying overrides. This aligns with `internal/service/scenario_service.go`.

<!-- Example Entry:
**Date:** YYYY-MM-DD
**Mistake:** [Description of the mistake, e.g., "Incorrectly implemented X functionality by doing Y."]
**File(s) Affected:** [List of files, if applicable]
**Resolution:** [Description of how the mistake was corrected, e.g., "Refactored Y to correctly implement X by Z. The key learning was ABC."]
--> 
