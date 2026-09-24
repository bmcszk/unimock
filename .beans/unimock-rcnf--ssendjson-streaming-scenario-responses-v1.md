---
# unimock-rcnf
title: SSE/NDJSON streaming scenario responses (v1)
status: todo
type: task
tags:
    - feature
created_at: 2026-09-24T13:19:20Z
updated_at: 2026-09-24T13:19:20Z
parent: unimock-xu0n
---

# SSE/NDJSON streaming scenario responses (v1)

Parent epic: unimock-xu0n (tier-1 differentiators, simple v1).

## What
A scenario can respond with a SERVER-GENERATED STREAM instead of a static body: SSE or NDJSON frames, interval-spaced, N events or hold-open.

## YAML shape (scenario gains optional key; mutually exclusive with `data`)
```yaml
scenarios:
  - method: GET
    path: /events
    status_code: 200
    stream:
      format: sse            # sse | ndjson
      interval_ms: 500       # delay between frames
      event_count: 10        # OR hold_open: true (exactly one)
      hold_open: false
      template: '{"n": {{index}}, "ts": "{{timestamp}}"}'   # {{index}} and {{timestamp}} substituted per frame
```

## Acceptance criteria
- [ ] Config parses + validates: format ∈ {sse, ndjson}; exactly one of event_count / hold_open; interval_ms ≥ 0 (default 0); mutually exclusive with data (config error otherwise)
- [ ] SSE frames: `data: <rendered template>\n\n` per event, Content-Type: text/event-stream, Cache-Control: no-cache
- [ ] NDJSON frames: rendered line + \n per event, Content-Type: application/x-ndjson
- [ ] Flush after every frame (http.Flusher); X-Accel-Buffering: no set automatically
- [ ] event_count mode: stream ends after N frames (connection closes)
- [ ] hold_open mode: streams until client disconnects (watch r.Context().Done())
- [ ] WriteHeader(200) BEFORE first flush
- [ ] NO compression, NO WebSocket, NO HTTP/2 push
- [ ] TDD: tests via httptest.NewServer + real http client (NEVER httptest.ResponseRecorder — its Flusher is a no-op)
- [ ] make check green; no //nolint; no lint-config changes

## Key files
- pkg/config/ (scenario schema + validation)
- internal/handler/ (stream writer — uni_handler writeBodyContent branch or new file)
