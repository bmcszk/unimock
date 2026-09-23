---
# unimock-uc44
title: README badges + GitHub rules (branch protection)
status: completed
type: task
priority: normal
created_at: 2026-09-23T23:00:23Z
updated_at: 2026-09-23T23:03:30Z
---

Epic: public-repo-maturity round 2 (user directive). (1) Badges on README: CI (PR+push workflows), Go Report Card, license, Go version, latest release. (2) gh rules: branch protection on master requiring PR + status checks. Constraint (HARD 2026-09-24): agent opens PR + verifies CI green, then STOPS — no self-merge. Branch protection settings need repo admin — document exact gh api commands for user to run, or apply if agent token has admin.

## Proof of Work
- 7 badges on README: pr.yml + push.yml workflow status (exact workflow names from live repo), Go Report Card, pkg.go.dev, release, MIT license, Go 1.26|1.27 — release + goreport badge endpoints verified 200
- gh rules found ALREADY ENFORCED via API (admin token): required PR, 3 status checks (PR tests, Push 1.26/1.27) strict, enforce_admins, no force-push/deletion
- docs/github-rules.md documents protection + merge policy: only bmcszk merges (HARD rule honored — NOT merged by agent)
- PR #37 all checks pass
Residual: merge pending with user
