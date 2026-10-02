---
# unimock-vhxx
title: 'Fix PR #44 review round: 4 MAJOR + 5 MINOR findings (_glm_first review)'
status: completed
type: task
priority: normal
created_at: 2026-10-02T15:46:47Z
updated_at: 2026-10-02T16:34:33Z
parent: unimock-xu0n
---

Root cause findings (inline on PR #44, ids 4167305977..4167311800):
MAJOR1 router.go:34 webhookDispatchTimeout=60s unsafe vs retry envelope
MAJOR2 dispatcher.go:249 signing key raw env string, not whsec_+base64 per Standard Webhooks spec
MAJOR3 stream_writer.go:105 hold_open+interval_ms:0 hot loop
MAJOR4 runtime scenario create/update (service.validateScenario) skips webhook/stream validation that LoadFromYAML enforces
MINOR5 dispatcher.go:76 blank-URL early return skips onComplete
MINOR6 dispatcher.go:247 Delivery.ID != webhook-id header
MINOR7 stream_writer.go:101 sleep not ctx-aware
MINOR8 docs/scenarios.md:579 SSE example double data: prefix
MINOR9 no stream_writer unit tests
Local: /home/blaze/github/bmcszk/unimock, branch feature/tier1-webhooks-streaming.
AC: all 9 fixed or explicitly rebutted; make check green; e2e green; lint 0; CI green; inline replies posted with fix SHAs.
Skills: go-dev, go-fluent-testing



## TRIAGE (brainer _glm_first, read-only @ 8a6998f)
All 9 findings REAL. 7 fix-as-suggested, 2 fix-different-way:
- MAJOR1: derive deadline via webhooks.DispatchTimeout(wh) (NOT clamp): MaxAttempts*10s+(MaxAttempts-1)*MaxMS+margin; also make backoff sleep ctx-aware (dispatcher.go:142). Defaults breach 60s: 3*10s+2*30s=90s.
- MAJOR2: fix-as-suggested — spec CONFIRMED (standard-webhooks §symmetric: whsec_ prefix + base64 StdEncoding decode before HMAC; svix-go reference). id.ts.payload construction stays.
- MAJOR3: validate error when HoldOpen && IntervalMS<=0 in StreamConfig.validate (don't floor in writer; interval_ms:0 + event_count burst stays legal).
- MAJOR4 KEYSTONE: move validators to pkg/model (model.Scenario.Validate() incl webhook+stream+data-exclusivity+newline rejection); pkg/config validateOneScenario delegates (YAML error strings stable); scenario_service.validateScenario calls scenario.Validate().
- MINOR5: defer/nil-guard onComplete before blank-URL early return (dispatcher.go:76).
- MINOR6: Delivery.ID = the deliveryID (one UUID per delivery), header webhook-id per-attempt fresh (as now); record deliveryID in ring record.
- MINOR7: ctx-aware wait seam in streamLoop: select ctx.Done()/time.After; keep injectable for tests.
- MINOR8: strip 'data: ' from doc example; reject \n\r in templates (in shared validator).
- MINOR9: internal/handler/stream_writer_test.go table-driven, injected now/wait, httptest.ResponseRecorder.
ORDER: 7→(2+6+5)→4→(3+8)→1→9. Risks: MAJOR4 must keep YAML error strings byte-compatible (config tests assert); MAJOR1 needs upper ceiling decision — covered by MAJOR4 bounds in validateRetries.
