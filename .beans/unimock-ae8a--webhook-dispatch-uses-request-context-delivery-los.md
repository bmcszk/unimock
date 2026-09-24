---
# unimock-ae8a
title: Webhook dispatch uses request context — delivery lost when client finishes reading before async fire
status: in-progress
type: bug
priority: normal
created_at: 2026-09-24T21:47:39Z
updated_at: 2026-09-24T22:07:17Z
parent: unimock-xu0n
---

## Evidence (E2E round, 2026-09-24)
TestWebhook_FiresOnScenarioMatch (tests/e2e/webhook_e2e_test.go, uncommitted): dispatcher log 'webhook delivery failed ... error: context canceled' — receiver saw 0 calls.
Root cause: internal/router/router.go:218 passes req.Context() into Dispatcher.Dispatch; handler returns as soon as the client reads the scenario response, canceling the context and killing the async delivery goroutine (internal/webhooks/dispatcher.go deliver loop aborts on ctx cancel). Any client that closes/disconnects fast loses the webhook — not just tests.
Fix direction (triage to confirm): detach dispatch lifetime from request — context.WithoutCancel(req.Context()) wrapped with dispatcher-owned timeout, or Background+timeout in maybeDispatchScenarioWebhook.
Streaming E2E file (streaming_e2e_test.go) never written — pi run died ~21:13 UTC+2 mid-contract; only webhook tests + DSL helpers landed.

## TRIAGE 2026-09-24 (brainer, read-only, verified @ feature/tier1-webhooks-streaming)

VERDICT: CONFIRMED — root cause as claimed. Reproduced live (go test ./tests/e2e -run TestWebhook_FiresOnScenarioMatch → FAIL, 'error: context canceled', receiver 0 calls).

### Root cause (verified file:line)
- internal/router/router.go:218 — Dispatch(req.Context(), ...) in maybeDispatchScenarioWebhook (:205); only production call site
- Async goroutine inherits req ctx (dispatcher.go:80 → deliver :111 → runAttempts :123, shouldAbort :149, NewRequestWithContext :203, canceledDelivery :180-188)
- Handler return → ctx canceled → delivery killed. Fast/disconnecting clients silently lose webhooks. Production bug, not test artifact.
- go.mod go 1.26.0 → context.WithoutCancel viable; toolchain go1.27.1
- DO NOT touch router.go:282 WriteStream(req.Context()) — streaming keeps request lifetime

### Fix shape
dispatchCtx, cancel := context.WithTimeout(context.WithoutCancel(req.Context()), webhookDispatchTimeout); defer cancel(); Dispatch(dispatchCtx, ...)
Timeout ≥ worst-case retry envelope (attempts×(10s client + 30s backoff)); suggest 60s floor.

### Acceptance criteria
- [ ] TestWebhook_FiresOnScenarioMatch passes
- [ ] All webhook e2e tests pass (retries et al.)
- [ ] Unit test: dispatch survives caller-ctx cancel (receiver still hit)
- [ ] Unit test: dispatchCtx timeout bounds delivery (no orphan goroutine)
- [ ] go vet + golangci-lint clean; go test ./... green
- [ ] router.go:282 WriteStream untouched (still req.Context())
- [ ] No other Dispatch call sites changed

### Skills for implementation
go-dev, go-fluent-testing (e2e assertions style)

### Subtasks (~10 min each)
1. const webhookDispatchTimeout + WithoutCancel+WithTimeout at router.go:218; e2e webhook tests
2. Unit test: caller cancels ctx post-Dispatch → receiver hit, delivery non-canceled
3. Unit test: dispatchCtx timeout bounds hanging receiver
4. Full gates + commit + close (main session)
