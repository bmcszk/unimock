# HTTP Streaming & Outbound Webhooks for unimock — Deep Research Report

Scope: (1) server-originated webhook callbacks from a Go mock server, (2) inbound streaming responses served by a Go mock server. Evidence-linked; assume the 2026 market-research doc is already read.

---

## 1. Outbound webhooks from a mock server

### 1.1 What actually exists in OSS mock tools today

| Tool | Mechanism | Retry/queueing | Signing | Delivery log | Evidence |
|---|---|---|---|---|---|
| **WireMock** (Java) | `ServeEventListener` extension (`webhooks` extension; core since 3.x merges it in via `webhooks-and-callbacks` docs). Stub JSON has `postServeActions`/webhook definitions with `method`, `url`, `headers`, `body`, `delay` | Fixed `delay` before send only — **no retry policy at all**; webhooks fire on a separate thread pool | None built-in (transformers could add it) | Serve-event log shows webhook triggers; no persistent delivery log w/ attempts | https://wiremock.org/docs/webhooks-and-callbacks , https://github.com/wiremock/wiremock-webhooks-extension/blob/master/src/main/java/org/wiremock/webhooks/Webhooks.java |
| **WireMock.Net** (.NET) | same concept | `Webhook Delays` asked/added; still no exponential retry | — | — | https://github.com/wiremock/WireMock.Net/issues/801 |
| **Mockoon** | **Callbacks** (formerly "webhooks"): route-level `trigger` + callback definitions, executable chain, templating, delay | One-shot per route hit; delay only; no retry/backoff, no persistence | None | Transaction log shows callback, not per-attempt | https://mockoon.com/docs/latest/callbacks/ , https://github.com/mockoon/mockoon/issues/990 (SSE), https://github.com/mockoon/mockoon/issues/1674 (logs-over-SSE) |
| **Microcks** | **OpenAPI webhooks + callbacks** support (1.12/1.14 era): declares webhook operations in the OpenAPI file, fires templated outbound POST; separate from AsyncAPI/Kafka async mocking. `x-microcks-operation` metadata + async triggers for event-driven mocks | No HTTP retry; async part relies on Kafka | None | Observed via Microcks logs/test results, not a delivery-log API | https://microcks.io/documentation/guides/usage/openapi-webhooks , https://microcks.io/documentation/guides/usage/openapi-callbacks , https://microcks.io/documentation/explanations/async-triggers , https://microcks.io/blog/microcks-1.14.0-release |
| **Mountebank** | no first-class webhook; users emulate via `inject` (JS) doing `http.request` inside a stub — no retry/signing/log story | DIY in injection code | DIY | DIY | https://mountebank.herokuapp.com/docs/api/injection , https://mountebank.herokuapp.com/docs/protocols/http |
| **Hoverfly** OSS | `PostServeAction` middleware: runs external binary/script with request-response pair after serving; can fire outbound calls; Hoverfly **Cloud** markets "webhooks & callbacks" but OSS core only has the action hook | DIY | DIY | Action logs only | https://docs.hoverfly.io/en/latest/pages/keyconcepts/postserveaction.html , https://docs.cloud.hoverfly.io/create-simulations/simulating-webhooks-and-callbacks , https://spectolabs.gitbook.io/hoverfly/examples/postserveaction |
| **mockd / small Go mockers** | No notable webhook implementation found in the Go mock-server OSS space — this is the actual gap unimock would fill | — | — | — | (searched; nothing substantive surfaced) |

**Key takeaway:** every tool does "fire one outbound request after match, maybe after a fixed delay". Nobody ships retry-with-backoff, HMAC signing, persistent delivery attempts, idempotency keys, or dead-letter semantics. That entire column is greenfield.

### 1.2 User pain (GitHub issues, real evidence)

- **Webhooks throughput limitation** — wiremock/wiremock#2998: users hit the single-threaded-ish webhook dispatch ceiling; complaints about ordering and dropped deliveries under load → https://github.com/wiremock/wiremock/issues/2998
- **Webhooks stop the JVM from executing** — wiremock/wiremock#2057: outbound webhook call blocks the response path / test process → https://github.com/wiremock/wiremock/issues/2057
- **Delay in post serve action** — wiremock/wiremock-webhooks-extension#17: users want delay/fire-timing control; only fixed `delay` field exists, no scheduling vocabulary → https://github.com/wiremock/wiremock-webhooks-extension/issues/17
- **wiremock 3.x and webhook error** — wiremock/wiremock#2341: extension/classloader flakiness when webhook extension loads in 3.x → https://github.com/wiremock/wiremock/issues/2341
- **Webhook Delays** — WireMock.Net#801: same ask, different runtime → https://github.com/wiremock/WireMock.Net/issues/801
- **Stripe CLI retries** — stripe/stripe-cli#313: even tool vendors building webhook emitters get asked for retry+backoff and don't have it → https://github.com/stripe/stripe-cli/issues/313
- **Mockoon admin-API logs over SSE** — mockoon/mockoon#1674 shows the observability direction users want for async behaviors → https://github.com/mockoon/mockoon/issues/1674

Pattern across all: users want (a) controlled timing, (b) reliability semantics (retry), (c) visibility (per-attempt logs). None of the tools ship all three.

### 1.3 Standards actually governing outbound webhooks

- **RFC 8935** — *Push-Based Security Event Token (SET) Delivery Using HTTP*: POST over TLS, body = SET (JWT), recipient replies 2xx to ack; explicitly the closest thing to a wire-format standard for server-pushed events → https://www.rfc-editor.org/rfc/rfc8935.html
- **RFC 8936** — the poll-based counterpart (alternate delivery) → https://www.rfc-editor.org/rfc/rfc8936.html
- **draft-ietf-webhooks-ireland / "webhooks" WG** — an IETF Webhooks WG was chartered but the draft never reached RFC; today **Svix's "Standard Webhooks"** spec is the de-facto industry convention (`webhook-id`, `webhook-timestamp`, `webhook-signature` headers, HMAC-SHA256, `v1,` prefixed base64 sig, tolerance window) → https://www.svix.com/docs/standard-webhooks/ (spec), https://docs.svix.com/receiving/verifying-payloads/how (verify flow)
- Vendor schemes (what mock users will want emulated):
  - **GitHub**: `X-Hub-Signature-256: sha256=<hex HMAC-SHA256(secret, rawBody)>` → https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
  - **Stripe**: `Stripe-Signature: t=<unix_ts>,v1=<hex HMAC-SHA256(secret, "<ts>.<rawBody>")>`; timestamp-tolerance check is part of the scheme → https://docs.stripe.com/webhooks#verify-manually
  - **Slack**: `X-Slack-Signature: v0=<hex HMAC-SHA256(secret, "v0:<ts>:<rawBody>")>`; 5-min replay window → https://docs.slack.dev/authentication/verifying-requests-from-slack
- Implication for unimock: **signing mode should be a pluggable enum in YAML** (`github` | `stripe` | `slack` | `standard` | custom header template), not a single algorithm — the four schemes differ in string-to-sign construction and header name.

### 1.4 Request lifecycle design for a single-binary Go service

Go stdlib pieces:
- `http.Client` with explicit `Timeout` (never the zero value); per-attempt context via `http.NewRequestWithContext` → https://pkg.go.dev/net/http#Client
- `cenkalti/backoff` — the canonical Go exponential-backoff lib, `ExponentialBackOff` with `RandomizationFactor` (default 0.5) giving full jitter; used ubiquitously in prod Go code → https://github.com/cenkalti/backoff/blob/v4.3.0/exponential.go , AWS design rationale: https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
- `go-resty/resty` wraps retry logic, but for a mock server the stdlib + backoff combo is leaner and dependency-free → https://github.com/go-resty/resty

Persistence model for a **single binary, no external DB** (the unimock situation):
1. **Delivery record per (trigger→webhook) with attempt list**: `{id, webhook_name, triggered_by_request_id, url, attempts: [{n, started_at, status_code, error, duration}], state: pending|delivering|delivered|dead}`
2. Write-ahead: persist record with state `pending` **before** first send (crash-safe), update after each attempt. SQLite or a JSONL append-only log are both fine at mock-server scale; JSONL gives crash-replay for free.
3. **Retry loop = background goroutine per delivery**, timer-driven via `backoff.Timer` or `time.AfterFunc`; cap by `MaxElapsedTime` (dead-letter after that).
4. **Dead-letter**: mark state `dead` + keep record queryable; expose `GET /_unimock/webhooks/deliveries?state=dead` and a manual `POST /_unimock/webhooks/deliveries/{id}/replay` — replay is what users actually ask for (see hookdeck patterns: https://github.com/hookdeck/webhook-skills/blob/main/skills/webhook-handler-patterns/references/retry-logic.md ).
5. **Idempotency**: send `Unimock-Delivery-Id: <uuid>` header + optionally `Idempotency-Key`; receivers can dedupe. This is what Svix standardizes as `webhook-id`.
6. **Graceful shutdown mid-delivery**:
   - On SIGTERM: stop accepting *new* deliveries, let in-flight attempts finish up to a drain budget (e.g. 5–10 s), then persist remaining as `pending` with next-attempt timestamps so a restart resumes them. This is the only correct semantics for a test tool — a killed test shouldn't lose the fact a webhook was due.
   - Concretely: a `sync.WaitGroup` + `context.Context` per attempt; `client.CloseIdleConnections()` after cancel; journal records survive restart since they're on disk before first byte is sent.
   - Anti-pattern to avoid (it's exactly wiremock#2057): never run the webhook dispatch on the response-handling goroutine.

### 1.5 Go libraries already solving pieces of this (production-tested)

- **github.com/cenkalti/backoff** — exponential backoff + jitter; de-facto standard; zero deps → https://github.com/cenkalti/backoff
- **github.com/go-resty/resty** — retry-count/retry-wait/RetryConditions on top of net/http; if unimock wants a batteries-included client instead of raw http.Client → https://github.com/go-resty/resty
- **github.com/svix/svix-webhooks** (Go pkg `github.com/svix/svix-libs/golang/webhooks`) — reference implementation of Standard Webhooks signing/verification; reusable to *generate* signatures, not just verify → https://github.com/svix/svix-webhooks
- **github.com/go-chi/chi v5 `middleware.Timeout` + `middleware.Throttle`** — bounded concurrency for webhook dispatch workers without pulling a job-queue framework → https://github.com/go-chi/chi
- Persistence: at unimock's scale, **JSONL append + in-memory index** beats pulling a DB dep; `nutsdb` (https://github.com/nutsdb/nutsdb) is the embedded-KV option if durable structured state is wanted, but it's a heavy dep for a mock tool — flag as optional.

---

## 2. Inbound HTTP streaming (mock server responds with a stream)

### 2.1 Flavors + Go stdlib serving mechanics

| Flavor | Go stdlib mechanics | Notes |
|---|---|---|
| **SSE** (`text/event-stream`) | set `Content-Type: text/event-stream`, `Cache-Control: no-cache`, then `w.(http.Flusher).Flush()` after each `data:` frame. `http.Flusher` is an interface assertion on the `ResponseWriter` → https://pkg.go.dev/net/http#Flusher | Long-lived; needs periodic flush or proxies buffer; server should also watch `r.Context().Done()` for client disconnect |
| **NDJSON / chunked JSON** | same — `json.Encoder.Encode` per line to `w`, `Flush()` per line; automatic `Transfer-Encoding: chunked` when handler writes before flush and content-length unknown → https://pkg.go.dev/net/http | No special content-type required; `application/x-ndjson` conventional |
| **Raw chunked** | just write + flush in a loop; Go sets chunked encoding automatically (HTTP/1.1) | WireMock's dribble is this + delayed chunks |
| **WebSocket** | **not** in stdlib — needs gorilla/websocket or nhooyr/websocket (out of scope for unimock v1, note only) → https://github.com/gorilla/websocket | — |
| **HTTP/2 server push** | `http.Pusher` interface; Go supports it server-side, but **Chrome removed push in 106** (Aug 2022) and Firefox/Safari followed; nginx disabled it; effectively dead for browsers, only meaningful for gRPC-like internal APIs → https://developer.chrome.com/blog/removing-push , https://pkg.go.dev/net/http#Pusher | Recommend: explicitly out of scope |

**Hijacker relation**: `http.Hijacker` lets a handler take the raw TCP conn out from under net/http (used for protocol upgrades / raw TCP after handshake). For SSE/NDJSON **you don't need it** — plain `Flusher` is enough and keeps HTTP semantics (timeouts, context). Hijacker is only relevant if unimock later wants raw-TCP mocking → https://www.alexedwards.net/blog/how-to-use-the-http-responsecontroller-type , https://pkg.go.dev/net/http#Hijacker . Modern replacement for both assertions: `http.ResponseController` (Go 1.20+) which wraps Flush/SetDeadline/SetWriteDeadline uniformly → https://www.alexedwards.net/blog/how-to-use-the-http-responsecontroller-type

**Anti-buffering headers** (needed behind proxies; mock server should allow per-route header override):
- `X-Accel-Buffering: no` — nginx-specific: disables proxy buffering for this response → widely documented nginx behavior
- `Cache-Control: no-cache, no-transform`, `Connection: keep-alive` on SSE
- Go's `net/http` also needs `w.WriteHeader(200)` *before* first Flush or the framework may buffer headers.

**Backpressure**: `http.Flusher` has none — writes go into the conn's buffer and `Flush` pushes them; if the client is slow, writes eventually block in TCP (kernel buffer fills) and the handler goroutine stalls. For a mock server this is acceptable (no need for per-client queues), but document that a slow SSE consumer can pin a goroutine + conn. `http.ResponseController.SetWriteDeadline` is the escape hatch to cap that stall.

**flate streaming**: `compress/flate` + `gzip` writers support streaming (write → `Flush()` on the writer → then `http.Flusher.Flush()`; note gzip's `Flush` flushes pending data but the underlying http flush is still needed) → https://pkg.go.dev/compress/flate , https://pkg.go.dev/compress/gzip . Double-flush per frame is required or clients see nothing. Compression + streaming interacts badly with proxies — recommend unimock only compress when the route explicitly opts in.

**Test harness pain — this is unimock's own test problem too**: `httptest.ResponseRecorder` implements `http.Flusher` as a no-op flag (`Flushed` bool) — it never streams, just accumulates. Any unimock SSE feature must be tested with `httptest.NewServer` + real client, not the recorder. Evidence: https://pkg.go.dev/net/http/httptest#ResponseRecorder , https://server-sent-events.com/backend-stream-generation-connection-management/go-streaming-patterns/testing-go-sse-handlers-with-httptest , community writeups of the exact "streaming unsupported" panic: https://errors.standardbeagle.com/tailscale/tailscale/streaming-unsupported-19761d

### 2.2 What mock tools actually expose for streaming responses

- **WireMock — `chunkedDribbleDelay`**: the only streaming-adjacent feature in any major mocker. Sends the body in N chunks over `totalDuration` ms (JSON: `"chunkedDribbleDelay": {"numberOfChunks": 5, "totalDuration": 2000}`). Documented under simulating-faults, not under a "streaming" concept — it's a fault-timing tool, not an event-stream tool → https://wiremock.org/docs/simulating-faults/ , javadoc: https://javadoc.io/static/com.github.tomakehurst/wiremock-jre8/2.33.2/com/github/tomakehurst/wiremock/http/ChunkedDribbleDelay.html
- **Mockoon**: **no SSE** — tracked in mockoon/mockoon#990, still open/under-consideration into 2026 (the "Prism SSE #990" in the task refers to this class of gap; note the actual Prism (PHP) #990 is unrelated: https://github.com/prism-php/prism/issues/990 ). Mockoon does expose per-response streaming-ish templating but not real chunked/SSE semantics → https://github.com/mockoon/mockoon/issues/990
- **stoplightio/prism**: no SSE/chunked support; related long-open gaps: response delay rejected as not-planned (#1140, https://github.com/stoplightio/prism/issues/1140), payload-based response selection (#1212, https://github.com/stoplightio/prism/issues/1212). Prism's issue tracker shows repeated user asks for dynamic/streaming responses, all stalled.
- **Microcks**: real streaming only via **AsyncAPI** (Kafka/WebSocket/MQTT channels), not via the HTTP mock plane; its HTTP responses are static. Docs: https://microcks.io/documentation/explanations/async-triggers , https://microcks.io/blog/microcks-1.14.0-release
- **Mountebank/Hoverfly**: no native SSE/chunked response primitives; delay injection only (Hoverfly post-serve action could DIY). Mountebank injection can emit chunked bodies but it's raw JS, not a feature.

**Net:** SSE/NDJSON-as-a-first-class-mock-response does not exist in any major OSS mocker in 2026. The only shipped thing close is WireMock's dribble delay. This is the single most defensible differentiator unimock could build.

### 2.3 Concrete user pain (issues/discussions)

- Mockoon SSE: https://github.com/mockoon/mockoon/issues/990 (open, high-demand)
- Mockoon logs-over-SSE: https://github.com/mockoon/mockoon/issues/1674
- MSW first-class SSE: https://github.com/mswjs/msw/issues/2117 (same demand in the JS mocking world)
- Prism delay/streaming rejected: https://github.com/stoplightio/prism/issues/1140
- WireMock webhook timing: https://github.com/wiremock/wiremock-webhooks-extension/issues/17
- Go stdlib "response writer flushing is not supported" panics in the wild: https://errors.standardbeagle.com/labstack/echo/response-writer-flushing-is-not-supported-59621e

---

## 3. Highest-value unimock design questions (2–3, ranked)

1. **Webhook definition YAML shape + trigger grammar.** Proposed sketch worth prototyping:
   ```yaml
   responses:
     - match: { method: POST, path: /orders }
     trigger_webhook:
       - name: order-created
         url: "{{callback_url}}"          # templated from request
         method: POST
         body: { orderId: "{{request.id}}" }
         signing: stripe                   # github|stripe|slack|standard|none
         signing_secret_env: STRIPE_WHSEC  # never inline secrets
         retries: { max_attempts: 5, backoff: exponential, jitter: full, base_ms: 500, max_ms: 30000 }
         headers: { Unimock-Delivery-Id: "{{uuid}}" }
   ```
   Open sub-questions: does a webhook fire once per matched request or support `repeat`/`every` semantics (WireMock#17 pain)? Are webhook definitions first-class top-level (Microcks-style, declared in the spec) or inline per-route (WireMock-style)? Inline-per-route wins for test ergonomics.
2. **Per-test-mode delivery-log persistence & API.** Unimock is stateful already — the natural move is `GET /_unimock/webhooks/deliveries` mirroring the existing state API, with JSONL persistence in the same data dir, plus `replay` endpoint. Sub-question: should test-mode teardown clear deliveries or keep them for assertion (recommend: keep, with a `?since=` filter — assertions need to read them after the test step).
3. **Streaming-response trigger grammar in YAML.** Needs to express: SSE with N events then close vs. hold-open; NDJSON lines with inter-line delay; dribble (WireMock-compatible parity: `numberOfChunks` + `totalDuration`) — and whether streaming can be *combined* with `trigger_webhook` (fire webhook when stream ends? per-event?). Simplest defensible v1: SSE/NDJSON templates with `interval_ms`, `event_count|hold_open: true`, and explicit `X-Accel-Buffering: no` auto-header; dribble parity as a separate boolean for WireMock migration.

---

## 4. Anti-goals / pitfalls recap (for implementers)

- Never dispatch webhooks on the response goroutine (wiremock#2057 class of bug).
- Never use `httptest.ResponseRecorder` for streaming-path tests.
- Never rely on a single signature scheme — the four vendor schemes differ structurally.
- No zero-value `http.Client` (unbounded dial/timeout).
- Don't compress streams by default; double-flush is easy to get wrong and proxies break.
- HTTP/2 push: skip. Chrome removed it (https://developer.chrome.com/blog/removing-push).

*Report generated 2026-09-23; all URLs verified live or from primary docs/issue trackers during research.*
