# GitHub Repository Rules

This document describes the governance rules enforced on this repository.
Branch protection is live on `master`; settings below are verified against
the GitHub API.

## Branch protection (`master`)

| Rule | Value |
|------|-------|
| Required PR | yes — no direct pushes (admins included, `enforce_admins`) |
| Required status checks | `Pull request tests`, `Push tests (1.26)`, `Push tests (1.27)` — strict (branch must be up to date) |
| Required approving reviews | 0 (checks gate instead; see review policy) |
| Force pushes / deletions | disabled |

## Review & merge policy

- **Only the repository owner (`bmcszk`) merges PRs.** Contributors and AI
  agents open PRs and ensure CI is green; merging is a human decision.
- PRs must pass all required checks before merge; squash-merge preferred
  (one commit per PR).
- The agent(s) working on this repo MUST NOT approve or merge their own PRs.

## CI gates

- **Pull request workflow**: lint (golangci-lint) + build + full e2e suite
  against Docker.
- **Push workflow**: lint + unit tests with coverage on Go 1.26 and 1.27
  (coverage badge published to `gh-pages` from Go 1.27 on master).
- **Security workflow** (weekly + on push/PR): govulncheck (call-graph
  reachable vulns), gosec (SAST), CodeQL, OpenSSF Scorecard — results to the
  Security tab. Docker images additionally Trivy-scanned on release.
- `make check` locally mirrors the lint+unit gates; `make test-e2e` mirrors e2e.

## Release process

1. Tag `vX.Y.Z` on master (chart version is stamped from the tag at publish).
2. `docker.yml` builds/pushes the image and publishes the Helm chart to
   `oci://ghcr.io/bmcszk/charts/unimock` with the tag version.
