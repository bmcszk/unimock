---
# unimock-jlc4
title: 'Modernize unimock: deps, cleanup, Go idioms'
status: completed
type: feature
priority: normal
created_at: 2026-09-22T19:28:44Z
updated_at: 2026-09-22T20:54:01Z
---

2026-09-22 user directive: 1) upgrade libs to latest stable, 2) drop claude/context7 remnants, 3) adopt beans tracking, 4) online Go best-practices research then ponytail refactor proposal (more Go, minimal, compact, readable, YAGNI).

## Summary of Changes
All 4 user directives done: (1) deps latest stable (go-restclient v0.2.0 etc., Go 1.26+), (2) claude/context7 artifacts removed, (3) beans adopted + all work tracked, (4) Go best practices researched (go.dev CodeReviewComments) and ponytail rounds 1+2 implemented.

## Proof of Work
PR (OPEN, CI 3/3 green): https://github.com/bmcszk/unimock/pull/28
Local gates re-verified after every round: make check -> vet 0, golangci 0, 287 unit PASS; make test-e2e -> 48 PASS (dockerized, Go 1.27).
Net repo: -370 lines; storage coverage 69.7% -> 89.8%; handler 80.0%.
Children: ghcb done, ykr4 done, jyjr done, ikln done, 7l4q done, 3zla scrapped (config-driven verdict), tvvr done.

Residual: none.
