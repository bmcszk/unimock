---
# unimock-7l4q
title: 'Ponytail refactor round 1: deletions + consumer-side interface'
status: in-progress
type: task
priority: normal
created_at: 2026-09-22T19:40:04Z
updated_at: 2026-09-22T19:58:56Z
---

Approved-scoped refactor per research in unimock-ikln: (1) delete internal/logger dead package; (2) shrink UniStorage 13 -> 6 methods (delete GetStrict/GetFlexible/UpdateStrict/UpdateFlexible/DeleteStrict/DeleteFlexible impls + dedicated tests, grep-driven, keep shared private helpers used by bool paths); (3) move interface to consumer pkg (internal/service), storage returns concrete type; (4) StorageError %v -> %w; (5) re-measure file sizes, no splits unless >1000 lines. Full make check + test-e2e after.
