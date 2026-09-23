# Contributing to Unimock

Thanks for your interest in contributing! This document covers the basics; the
detailed engineering standards live in [`docs/testing-guidelines.md`](docs/testing-guidelines.md)
and [`docs/pr-guidelines.md`](docs/pr-guidelines.md).

## Getting started

```bash
git clone https://github.com/bmcszk/unimock.git
cd unimock
make deps        # install dev tooling
make build       # build the binary
make test        # unit tests
make check       # full validation (build + vet + lint + tests) — run before every commit
```

## How we work

- **Branches, never direct pushes to master.** Fork or cut a branch, open a PR.
- **`make check` must pass** on every commit. CI runs the same gates on Go 1.26 and 1.27.
- **No `//nolint` and no lint-configuration changes** to make code pass — fix the code.
- **TDD when adding or changing logic**: test first, then implementation.
- **E2E tests** run against Docker (`make test-e2e`) using the fluent testing
  style described in [`docs/fluent-testing.md`](docs/fluent-testing.md).

## Pull requests

1. One logical change per PR.
2. Run `make check` locally — CI failures that a local run would have caught
   waste everyone's time.
3. Describe **what** changed and **why**; link the issue/bean if there is one.
4. Address review comments systematically — see the zero-tolerance tracking
   process in [`docs/pr-guidelines.md`](docs/pr-guidelines.md).

## Reporting bugs

Open a GitHub issue with:

- Unimock version (`docker image tag`, `git describe`, or Helm chart version)
- Minimal `config.yaml` reproducing the problem
- Expected vs actual behaviour, with `curl` transcripts or logs

## Reporting security issues

See [SECURITY.md](SECURITY.md). Please do not open public issues for
security vulnerabilities.

## License

By contributing you agree that your contributions are licensed under the
[MIT License](LICENSE).
