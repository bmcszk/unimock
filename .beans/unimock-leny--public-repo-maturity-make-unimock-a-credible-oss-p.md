---
# unimock-leny
title: 'Public-repo maturity: make unimock a credible OSS project'
status: completed
type: feature
priority: normal
created_at: 2026-09-23T19:34:32Z
updated_at: 2026-09-23T19:55:22Z
---

Session goal: lift unimock from working-code to credible public OSS repo. Research-backed gaps from docs/market-research-2026.md + user directives. Candidate work: research docs visibility (PR #31 pending merge), README polish, LICENSE/versioning consistency (Chart.yaml 1.0.0 vs pkg.go.dev v0.0.8), CI green, examples that run, community files (CONTRIBUTING/SECURITY/CODE_OF_CONDUCT), agent-friendliness (agents.md), roadmap derived from research priorities.

## Proof of Work
Epic complete. All 5 children completed:
- unimock-ewb0: research docs on master (PR #31)
- unimock-uq43: Chart version aligned to v0.0.8 (PR #32)
- unimock-qc1m: README verified live + XML fixture bug found & fixed (PR #33)
- unimock-rr92: CONTRIBUTING/SECURITY/CODE_OF_CONDUCT added (PR #34)
- unimock-gzly: AGENTS.md + examples audited live (PR #35)

End-to-end state on master (c451868):
- CI green on all 6 merged PRs; master push workflow: completed success
- Community files present at repo root; CONTRIBUTING gates match real Makefile
- README quick-start verified with docker run + curl (real execution)
- Example configs load with zero fixture errors; XML scenario serves
- Agent surface: AGENTS.md + scenario-precedence caveat in examples
Residual: README badges (CI/GoReport/license) not yet added — cosmetic, candidate for follow-up; repo otherwise credible-OSS baseline reached
