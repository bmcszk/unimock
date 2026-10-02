---
# unimock-xu0n
title: 'Epic: tier-1 webhooks + streaming (simple v1)'
status: completed
type: feature
priority: normal
tags:
    - epic
created_at: 2026-09-24T13:17:45Z
updated_at: 2026-09-24T18:23:27Z
---

# Epic: Tier-1 differentiators — outbound webhooks + streaming responses (simple v1)

Source: docs/webhooks-streaming-research-2026.md + docs/market-research-2026.md (both on master).

## Goal
Ship the two features no major OSS mock server has first-class, as MINIMAL v1 implementations:
1. Outbound webhook triggering from scenario matches — fire one POST after a scenario
   matches, with retry+backoff and an HMAC signature enum. NO TLS/cert handling.
2. Streaming responses (SSE + NDJSON) from scenarios — template frames with interval
   and count/hold-open. NO compression, NO WebSocket, NO HTTP/2 push.

## Explicitly out of scope (v1)
- Any certificate/TLS configuration handling
- WebSocket, HTTP/2 push, gzip streaming
- Webhook scheduling/repeat semantics (fire-once per match only)
- Signature: single HMAC-SHA256 standard scheme only (Standard Webhooks headers),
  no github/stripe/slack scheme emulation in v1
- Delivery log persistence to disk (in-memory + GET endpoint only)
- templating beyond {{request.<path>}} / {{uuid}} minimal set if complex

## Constraints
- make check green (vet+lint+unit); TDD; no //nolint; no lint-config changes
- Branch + PR; never push master
- Anti-goals from research: never dispatch webhooks on the response goroutine;
  explicit-timeout http.Client only; streaming tests use httptest.NewServer not ResponseRecorder
