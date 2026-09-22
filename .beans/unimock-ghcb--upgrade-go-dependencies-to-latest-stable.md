---
# unimock-ghcb
title: Upgrade Go dependencies to latest stable
status: completed
type: task
priority: normal
created_at: 2026-09-22T19:28:34Z
updated_at: 2026-09-22T19:35:27Z
parent: unimock-jlc4
---

go.mod deps are stale: go-restclient v0.0.9 (latest v0.2.0), chi v5.2.2 (v5.3.2), testify v1.10.0 (v1.12.1), antchfx jsonquery/xmlquery/xpath outdated, golang.org/x/* months old. Bump all direct deps to latest stable, go mod tidy, full test suite green. Refs: modernization epic.

## Summary of Changes
All direct deps at latest stable (go-restclient v0.2.0, testify v1.12.1, chi v5.3.2, antchfx latest, x/* latest); go directive 1.26.0 forced by x/net+text; CI + Dockerfile moved to 1.26/1.27. Linter fixes: dropped middleware.RealIP (deprecated, spoofing), internal/errors -> internal/errs, wg.Go in test.

## Proof of Work
make check   -> 0 lint issues, 287 unit tests PASS (race+cover)
make test-e2e -> docker build OK (golang:1.27-alpine), 48 e2e tests PASS

Residual: none.
