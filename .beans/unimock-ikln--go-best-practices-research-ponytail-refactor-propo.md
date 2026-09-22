---
# unimock-ikln
title: Go best-practices research + ponytail refactor proposal
status: completed
type: task
priority: normal
created_at: 2026-09-22T19:28:34Z
updated_at: 2026-09-22T19:40:04Z
parent: unimock-jlc4
---

Learn current Go good practices online (effective_go-style modern idioms, golangci config, 2025-2026 recommendations), then propose a ponytail (lazy/minimal/YAGNI) refactor plan: more Go, more minimal, more compact, more readable. Research first, proposal to user for approval before touching code. Refs: modernization epic.

## Summary of Changes
Researched go.dev CodeReviewComments + Effective Go (go.dev fetched 2026-09-22) and audited unimock against it. Proposal delivered to user in chat; implementation bean drafted (unimock-ref1).

## Proof of Work
Sources: https://go.dev/wiki/CodeReviewComments (full text), go.dev/dl (version check). Findings: (1) internal/logger dead pkg 0 importers; (2) UniStorage iface 13 methods, 6 Strict/Flexible twins have 0 non-test callers, service uses only bool-flag 6; iface in producer pkg (wiki: belongs in consumer); (3) 20 bool-flag params; (4) uni_storage.go 935 lines shrinks under ceiling after twin deletion; (5) errs StorageError wraps with %v not %w.

Residual: implementation pending user approval.
