---
# unimock-qc1m
title: README first-impression pass
status: completed
type: task
priority: normal
created_at: 2026-09-23T19:35:39Z
updated_at: 2026-09-23T19:44:40Z
---

Epic: unimock-leny. Quick-start copy-paste works, feature claims accurate vs code, links valid. Gate: every README command runs green in clean checkout.

## Progress (partial — XML fixture bug split out)
- README quick-start verified LIVE: docker run + POST/GET /api/users green
- All 3 example configs parse as YAML; config-unified + scenarios.yaml load in server
- Feature claims checked: no SSE/webhook overpromises in README
- Found bug: plain XML bodies treated as fixture syntax (logged error + fallback) — fixed in PR #33, merged, e2e 48 green
- Missing community files confirmed: SECURITY.md, CODE_OF_CONDUCT.md, CONTRIBUTING.md absent (next: unimock-rr92)
Residual: none for this bean — README itself accurate; examples now run clean
