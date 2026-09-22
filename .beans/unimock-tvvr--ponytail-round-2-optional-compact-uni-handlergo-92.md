---
# unimock-tvvr
title: 'Ponytail round 2 (optional): compact uni_handler.go 920-line dispatch'
status: completed
type: task
priority: normal
created_at: 2026-09-22T20:26:16Z
updated_at: 2026-09-22T20:49:15Z
parent: unimock-jlc4
---

uni_handler.go 48 funcs / 920 lines, switch on req.Method at :807. Candidate for table-driven dispatch or cohesive splits. Punt from round 1 — works, tested. Do only if it grows or changes.

## Summary of Changes
- [x] read uni_handler.go end-to-end; method switch kept (idiomatic)
- [x] applyRequest/ResponseTransformations -> shared runTransforms generic helper (pure move)
- [x] dead GetConfig() deleted (zero callers repo-wide incl. docs)
- [x] no file split: 902 lines < 1000 ceiling

## Proof of Work
pi contract run (_minimax_first) + orchestrator re-gate: make check -> vet 0, golangci 0, 287 unit PASS; handler coverage 80.0%. Commit 729f6d1 pushed.

Residual: none.
