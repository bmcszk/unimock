# Architectural and Design Decisions

This document records significant architectural and design decisions made throughout the project lifecycle. Each entry should include the decision, the rationale behind it, and the date it was made.

---

## Decision Log

**Date:** 2025-05-28
**Decision:** The `pkg/model.Scenario` struct will maintain a flat structure.
**Rationale:** The `Scenario` model is used for both matching incoming requests (based on `RequestPath`, which includes method and path) and defining the complete override response (including `StatusCode`, `ContentType`, `Data`/Body, `Location` header). A flat structure simplifies its usage in `internal/handler/mock_handler.go` and `internal/service/scenario_service.go` and directly reflects the fields needed for these operations. Refactoring into nested sub-structs (e.g., `ScenarioRequest`, `ScenarioResponse`) was found to add unnecessary complexity and did not align with the direct way these fields are utilized.

(No decisions logged yet)

<!-- Example Entry:
**Date:** YYYY-MM-DD
**Decision:** Adopted [Technology/Pattern X] for [Specific Purpose Y].
**Rationale:** [Brief explanation of why this decision was made, alternatives considered, and trade-offs.]
--> 
