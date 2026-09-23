---
# unimock-rr92
title: 'Community files: CONTRIBUTING, SECURITY, CODE_OF_CONDUCT'
status: completed
type: task
priority: normal
created_at: 2026-09-23T19:35:39Z
updated_at: 2026-09-23T19:48:14Z
---

Epic: unimock-leny. CONTRIBUTING refs make check + docs/pr-guidelines.md. Gate: files exist, links valid, CONTRIBUTING steps pass.

## Proof of Work
- CONTRIBUTING.md: real gates verified against Makefile (make check runs vet+lint+unit; e2e via make test-e2e), references existing docs/pr-guidelines.md + docs/fluent-testing.md, no //nolint rule stated
- SECURITY.md: threat model honest (test tool, no auth on /_uni/*), private vuln reporting via GitHub advisories URL (valid path format)
- CODE_OF_CONDUCT.md: Contributor Covenant 2.1
- All 3 links checked against repo files (LICENSE, docs/*)
- PR #34 checks: Pull request tests pass, Push tests 1.26+1.27 pass; merged squash
Residual: none
