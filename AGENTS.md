# AGENTS.md

Guidance for AI coding agents (Claude Code, Codex, Cursor, pi, etc.) working
in this repository, and a schema primer for agents that **use** unimock to
mock HTTP dependencies.

## Working in this repo

- **Validation gate:** `make check` (vet + golangci-lint + unit tests) must
  pass before you claim any change is done. E2E: `make test-e2e` (Docker).
- **No `//nolint` comments. No lint-config changes** to silence findings —
  fix the code.
- **Tests first** for logic changes (TDD). E2E tests use the fluent style in
  [`docs/fluent-testing.md`](docs/fluent-testing.md).
- **Branch + PR.** Never push to master. One logical change per PR.
- Config code lives in `pkg/config/`, HTTP handling in `internal/handler/`,
  storage in `internal/storage/`. Docs in `docs/`, PRDs in `docs/prds/`.

## Using unimock (YAML schema primer)

Unimock is configured with a single YAML file (`UNIMOCK_CONFIG`, default
`config.yaml`). Unified format has two top-level keys:

```yaml
sections:            # stateful CRUD resources
  <name>:            # section name = storage bucket
    path_pattern: "/api/users/*"   # * single segment, ** recursive
    body_id_paths: ["/id", "/user/id"]  # JSON pointers to extract the resource ID from the body
    header_id_names: ["X-User-ID"]      # ...or from headers (first match wins)
    return_body: true                # echo stored body on GET
    strict_path: false               # true = reject IDs accessed across different path patterns

scenarios:           # predefined responses — OVERRIDE storage on method+path match
  - uuid: "user-not-found"         # optional; auto-generated if empty
    method: "GET"
    path: "/api/users/999"         # exact match
    status_code: 404
    content_type: "application/json"
    headers:                       # optional response headers
      X-Trace: "abc"
    data: '{"error": "User not found"}'
```

Key semantics agents must know:

1. **Scenarios take precedence over stored resources.** A `POST /api/users`
   scenario means POSTs on that path return the canned response and the body
   is *not* stored — later GETs hit 404 unless a section stored it another way.
2. **ID extraction** creates implicit CRUD: after a successful POST/PATCH,
   GET/PUT/DELETE work on `path_pattern-prefix + /<id>`.
3. **`data` supports fixture references**: `@ ./file`, `< ./file`,
   `<@ ./file` (go-restclient compatible). Plain XML bodies are content —
   only `< `/`<@` prefixed data is treated as fixture syntax.
4. **Sections can nest**: `path_pattern: "/api/users/*/orders/*"` for
   sub-resources.
5. **Admin API** (for agent tooling): `GET/POST/DELETE /_uni/scenarios`,
   `GET /_uni/health`, `GET /_uni/metrics`. Go client:
   [`pkg/client`](pkg/client/).

## Environment

- `UNIMOCK_PORT` (default 8080), `UNIMOCK_LOG_LEVEL` (debug|info|warn|error),
  `UNIMOCK_CONFIG` (default `config.yaml`).
- Runnable examples: `examples/configs/` (verified to load with zero errors).
