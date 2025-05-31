# Architectural and Design Decisions

This document records significant architectural and design decisions made throughout the project lifecycle. Each entry should include the decision, the rationale behind it, and the date it was made.

---

## Decision Log

**Date:** 2025-05-28
**Decision:** The `pkg/model.Scenario` struct will maintain a flat structure.
**Rationale:** The `Scenario` model is used for both matching incoming requests (based on `RequestPath`, which includes method and path) and defining the complete override response (including `StatusCode`, `ContentType`, `Data`/Body, `Location` header). A flat structure simplifies its usage in `internal/handler/mock_handler.go` and `internal/service/scenario_service.go` and directly reflects the fields needed for these operations. Refactoring into nested sub-structs (e.g., `ScenarioRequest`, `ScenarioResponse`) was found to add unnecessary complexity and did not align with the direct way these fields are utilized.

**Date:** 2025-05-28
**Decision:** Adopted a dual-map approach for storing resources with multiple identifiers in `internal/storage/mock_storage.go` to support REQ-RM-MULTI-ID.
**Data Structures:**
1.  `resourceData map[string]model.Resource`: Stores the actual resource content. Keyed by a unique, internally generated primary ID (UUID). `model.Resource` will encapsulate the resource body, headers, status, etc.
2.  `idMappings map[string]string`: Maps various external identifiers (from path, configured headers, or body fields) to the internal primary ID in `resourceData`.
**Rationale:**
- **Single Source of Truth:** Resource data is stored once in `resourceData`, preventing duplication and ensuring consistency.
- **Flexible Identification:** `idMappings` allows a resource to be looked up via any of its associated external IDs.
- **Conflict Detection:** Ensures an external ID is not inadvertently mapped to multiple distinct resources by checking `idMappings` during creation.
- **Decoupling:** The internal primary ID decouples the physical storage from the multiple ways a resource can be identified externally.
- **Integration with Path:** While `resourceData` and `idMappings` are global, `model.Resource` will still store its canonical request path. The primary ID for a resource will be unique globally. Path-based lookups (e.g. GET /collection/) will still iterate relevant resources, and individual resource lookups (GET /collection/some-id) will use `some-id` in `idMappings`.

---

**Date:** 2025-05-31
**Decision:** Step 2 (Retrieve by Secondary ID, e.g., `GET /products?sku=...`) of PRD `REQ-E2E-COMPLEX-001` will be skipped in the E2E test `TestSCEN_E2E_COMPLEX_001_MultistageResourceLifecycle`.
**Rationale:** Unimock currently only supports resource retrieval via direct path identifiers (e.g., `url/{id}`). It does not support querying collections or filtering by secondary identifiers using query parameters like `?sku=...`. Therefore, this specific test step is untestable with the current Unimock capabilities and has been removed from the E2E HTTP request/response files and the corresponding Go test logic.

<!-- Example Entry:
**Date:** YYYY-MM-DD
**Decision:** Adopted [Technology/Pattern X] for [Specific Purpose Y].
**Rationale:** [Brief explanation of why this decision was made, alternatives considered, and trade-offs.]
--> 
