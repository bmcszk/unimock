---
# unimock-7l4q
title: 'Ponytail refactor round 1: deletions + consumer-side interface'
status: completed
type: task
priority: normal
created_at: 2026-09-22T19:40:04Z
updated_at: 2026-09-22T20:25:35Z
---

Approved-scoped refactor per research in unimock-ikln: (1) delete internal/logger dead package; (2) shrink UniStorage 13 -> 6 methods (delete GetStrict/GetFlexible/UpdateStrict/UpdateFlexible/DeleteStrict/DeleteFlexible impls + dedicated tests, grep-driven, keep shared private helpers used by bool paths); (3) move interface to consumer pkg (internal/service), storage returns concrete type; (4) StorageError %v -> %w; (5) re-measure file sizes, no splits unless >1000 lines. Full make check + test-e2e after.

## Summary of Changes
- [x] delete internal/logger (dead pkg)
- [x] UniStorage 13 -> 6 methods (twins deleted, -160 lines storage)
- [x] interface -> exported concrete *storage.UniStorage, service consumes it
- [x] StorageError + service wrapping %v -> %w
- [x] no file splits (789 < 1000 after deletions)

## Proof of Work
make check    -> vet 0, golangci 0 issues, 287 unit PASS (race+cover)
make test-e2e -> 48 PASS on dockerized server
storage coverage 69.7% -> 89.8%; repo net -339/+182 lines (a58eb4e)

Residual: round-2 candidates filed as drafts (bool-flag params, handler dispatch table).
