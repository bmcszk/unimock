---
# unimock-gzly
title: agents.md + runnable examples audit
status: completed
type: task
priority: normal
created_at: 2026-09-23T19:35:39Z
updated_at: 2026-09-23T19:54:13Z
---

Epic: unimock-leny. agents.md teaching YAML schema for agents; examples audited vs live server. Gate: doc examples run green, agents.md accurate.

## Proof of Work
- Live audit: 28 curl flows from docs/examples.md run against server; full CRUD + multi-ID extraction (sku+uuid) PASS on section config
- Found + documented: intercepting scenarios mean POST bodies are NOT stored (docs/scenarios.md confirms precedence) — examples.md annotated with config-note
- agents.md created: repo gates (make check/no //nolint/TDD/branch-PR) + full YAML schema primer (sections, scenarios, ID extraction, fixture syntax, precedence, admin API)
- Schema claims verified against pkg/client + docs/technical_endpoints.md + live runs
- PR #35 checks pass (PR tests, Go 1.26+1.27); merged
Residual: none
