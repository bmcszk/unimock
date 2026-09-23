---
# unimock-uc44
title: README badges + GitHub rules (branch protection)
status: in-progress
type: task
created_at: 2026-09-23T23:00:23Z
updated_at: 2026-09-23T23:00:23Z
---

Epic: public-repo-maturity round 2 (user directive). (1) Badges on README: CI (PR+push workflows), Go Report Card, license, Go version, latest release. (2) gh rules: branch protection on master requiring PR + status checks. Constraint (HARD 2026-09-24): agent opens PR + verifies CI green, then STOPS — no self-merge. Branch protection settings need repo admin — document exact gh api commands for user to run, or apply if agent token has admin.
