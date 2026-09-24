# Webhook trigger from scenario match (v1)

Parent epic: unimock-xu0n (tier-1 differentiators, simple v1).

## What
Scenario-level optional webhook fires ONE outbound HTTP request after a scenario matches.

## YAML shape (scenario gains optional key)
```yaml
scenarios:
  - method: POST
    path: /orders
    status_code: 201
    webhook:
      url: "http://callback.example/hook"   # required, must parse
      method: POST                          # optional, default POST
      headers: {}                           # optional extra headers
      body: '{"orderId": "{{uuid}}"}'       # optional; default = matched request body
      secret_env: MY_WHSEC                  # optional; env var name, Standard Webhooks HMAC
      retries:
        max_attempts: 3                     # default 3
        base_ms: 500                        # default 500
        max_ms: 30000                       # default 30000
```

## Acceptance criteria
- [ ] Config parses + validates: url must parse; method restricted to POST/PUT/PATCH; unknown fields rejected at load; secret NEVER inline (only env var name)
- [ ] Signing: secret_env set → Standard Webhooks headers (webhook-id: uuid, webhook-timestamp: unix, webhook-signature: `v1,` base64 HMAC-SHA256 over `<id>.<ts>.<body>`), secret read from env at dispatch time
- [ ] Dispatch is async (own goroutine; response path never blocks on it)
- [ ] http.Client with explicit Timeout (no zero-value client)
- [ ] Retry: exponential backoff full jitter (base*2^n capped at max_ms); stop on 2xx; max_attempts total
- [ ] Delivery log: in-memory ring buffer (cap 100) + `GET /_uni/webhooks/deliveries` returning JSON array {id, url, attempt, status_code|error, ts}
- [ ] NO TLS/cert handling anywhere
- [ ] TDD: unit tests for signer, backoff, ring buffer, config validation; handler test fires webhook against httptest.NewServer receiver
- [ ] make check green (vet + lint + unit); no //nolint; no lint-config changes

## Key files
- pkg/config/ (scenario schema + validation), internal/handler/ (dispatch + delivery endpoint), internal/webhooks/ or internal/service/ (dispatcher, signer, backoff, ring)
