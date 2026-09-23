---
# unimock-uq43
title: Align Chart.yaml version with module tags
status: completed
type: bug
priority: normal
created_at: 2026-09-23T19:35:39Z
updated_at: 2026-09-23T19:39:13Z
---

Epic: unimock-leny. Chart.yaml version+appVersion 1.0.0 vs module v0.0.8. Gate: helm chart version matches latest git tag scheme, CI green.

## Proof of Work
- Root cause: docker.yml overwrites Chart version+appVersion from git tag at publish, so committed values are placeholders — but they said 1.0.0 vs latest tag v0.0.8, misleading source installs
- Fix: Chart.yaml 0.0.8/0.0.8 + README tgz name fix
- PR #32 checks: Pull request tests pass, Push tests 1.26+1.27 pass
- gh pr merge 32 --squash: merged; master HEAD shows the fix
Residual: none (CI owns version at publish)
