---
# unimock-uc44
title: README badges + GitHub rules (branch protection)
status: completed
type: task
priority: normal
created_at: 2026-09-23T23:00:23Z
updated_at: 2026-09-23T23:20:56Z
---

Epic: public-repo-maturity round 2 (user directive). (1) Badges on README: CI (PR+push workflows), Go Report Card, license, Go version, latest release. (2) gh rules: branch protection on master requiring PR + status checks. Constraint (HARD 2026-09-24): agent opens PR + verifies CI green, then STOPS — no self-merge. Branch protection settings need repo admin — document exact gh api commands for user to run, or apply if agent token has admin.

## Proof of Work
- 7 badges on README: pr.yml + push.yml workflow status (exact workflow names from live repo), Go Report Card, pkg.go.dev, release, MIT license, Go 1.26|1.27 — release + goreport badge endpoints verified 200
- gh rules found ALREADY ENFORCED via API (admin token): required PR, 3 status checks (PR tests, Push 1.26/1.27) strict, enforce_admins, no force-push/deletion
- docs/github-rules.md documents protection + merge policy: only bmcszk merges (HARD rule honored — NOT merged by agent)
- PR #37 all checks pass
Residual: merge pending with user

## Round 2 (user: gha not configured, no security checks like go-restclient)
- security.yml ported from go-restclient: govulncheck + gosec SAST + CodeQL + OpenSSF Scorecard (publish_results for public badge); weekly cron + push/PR
- push.yml: coverage profile + gh-pages badge publish (Go 1.27/master, same script as go-restclient)
- README: +Security, +Scorecard, +coverage badges (now 10)
- Fix en route: scorecard-action@v2 tag unresolvable -> pinned v2.4.4 SHA (go-restclient's exact pin)
- All 9 checks pass (govulncheck, gosec, CodeQL, Scorecard, PR tests, Push 1.26/1.27, CodeQL scan, Scorecard result)
- github-rules.md documents the security gates
- NOT merged per HARD rule — awaiting bmcszk
