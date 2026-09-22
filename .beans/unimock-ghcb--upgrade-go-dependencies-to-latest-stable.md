---
# unimock-ghcb
title: Upgrade Go dependencies to latest stable
status: in-progress
type: task
priority: normal
created_at: 2026-09-22T19:28:34Z
updated_at: 2026-09-22T19:28:44Z
parent: unimock-jlc4
---

go.mod deps are stale: go-restclient v0.0.9 (latest v0.2.0), chi v5.2.2 (v5.3.2), testify v1.10.0 (v1.12.1), antchfx jsonquery/xmlquery/xpath outdated, golang.org/x/* months old. Bump all direct deps to latest stable, go mod tidy, full test suite green. Refs: modernization epic.
