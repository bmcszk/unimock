# Unimock — potential customer research synthesis (2026-09-22)

Sources: 2 research subagents (GitHub issue mining across wiremock/prism/mockoon/hoverfly
+ buyer-side pricing/adoption) + own searches (Postman State of API 2025, CircleCI 2026
delivery report, Docker State of App Dev 2025, protocol share stats, MCP/AI tooling,
Go-native competitor scan). Full evidence links in body below + companion file.

## Verdict up front
Unimock's existing core (stateful CRUD + auto ID extraction + YAML config + single Go
binary + Docker/Helm) hits the #1 validated feature gap and the #1 validated switching
trigger in the market. The winnable niche is "WireMock-class statefulness without JVM,
YAML-native, self-hosted, agent-friendly". The trap is competing on protocol breadth
(gRPC/GraphQL/Kafka) — that war is already lost to Microcks (CNCF) and mockd.

## 1. Who the customers are
- Frontend devs blocked on backend APIs → need stateful fake now (json-server crowd;
  they defect to whatever gives CRUD + filtering free).
- QA/platform teams killing flaky integration tests → staging is contention hell
  (Signadot/Google hermetic-env literature); mocks are the isolation tool.
- Go/non-JVM shops → resent needing a JVM for mocks; WireMock OOM/thread-leak reports
  are the trigger event (groups.google.com/g/wiremock-user/c/BxCBhCBATTA, issue #1346).
- Mid-market priced out of SV: WireMock Cloud free = 3 APIs/10 rps/NO statefulness;
  statefulness+K8s+versioning+MCP all enterprise-quote (median ~$15k/yr per Vendr);
  Parasoft ~$56k/yr; Traffic Parrot ~$10k+. No vendor has a mid tier. Mountebank dead
  (author quit 2024). stubby4j archived. This is the empty seat.

## 2. What they need (evidence-ranked)
1. Stateful CRUD with ID extraction — #1 gap everywhere: wiremock-state-extension
   exists because of it; Mockoon CRUD filtering took to v6.1 (#1047), custom IDs still
   open (#1547); Prism "Data Persistence" unbuilt for years. json-server wins frontend
   precisely here. → unimock's core feature is the differentiator.
2. YAML config-as-code — nobody major is YAML-native (WireMock JSON + YAML stub PR
   open since 2023 #2411; Mockoon JSON/GUI; Hoverfly hostile JSON #815 + needs an LSP
   to edit safely). Teams live in Git and want reviewable diffs.
3. Single binary, low memory, fast start — JVM tax complaints + "lightweight CLI"
   category exists because of it.
4. Correct concurrency on stateful mocks — WireMock scenario races under load
   (community.wiremock.io/t/16926353). Go sync is a real quality argument.
5. Great no-match diagnostics — matcher footguns are a permanent WireMock support
   topic; machine-readable near-miss reports would serve humans AND agents.
6. Runtime state reset + isolation for shared instances — Mockoon needed
   POST /mockoon-admin/state/purge (#1334); WireMock global reset wipes other tests'
   stubs. Scoped reset = table stakes for shared dev instances.
7. Validation at load (fail fast with precise errors) — Hoverfly's silent mismatch
   then match-time panic is the counterexample.
8. Hot reload of config (Mockingjay/mockserverx already advertise it).
9. OpenAPI import (even lossy/partial) — WireMock OSS ignores specs entirely; stub-
   per-endpoint file explosion is documented pain (Twilio Verify = 53 hand-written
   stubs). Not required for the niche but removes the biggest onboarding objection.
10. AI/agent surface — WireMock Cloud's MCP + agent skills are Cloud-paid only;
    Imposter ships agent-skills; mockd exposes 18 MCP tools; Specmatic has an MCP
    server. A local-first OSS MCP layer (reload_config, reset_state,
    list_unmatched_requests, verify_request) + an agents.md teaching the YAML schema
    is cheap and differentiating.

## 3. Market facts
- REST = 78-83% of APIs in use (Postman SoA 2025 N=49k; ProgrammableWeb tracker);
  GraphQL 41% using; gRPC 31% but fastest-growing new-project choice (16%→24% since
  2022). REST mocking is NOT a shrinking niche at the edge; internal gRPC is where
  growth is.
- 64% of devs now work in non-local environments (Docker 2025) → deployable/shareable
  mocks matter more every year; unimock's Helm chart is ahead of most OSS peers here.
- CircleCI 2026: integration is THE bottleneck; AI-written code raises failure rates;
  MTTR 72min. Anything that makes contract drift visible early sells.
- Star benchmarks: Mockoon 8.4k, WireMock 7.4k, Prism 5.0k, Hoverfly 2.5k,
  Microcks 2.0k (CNCF incubating 2026-05), unimock 1. Category winners cap at
  5–8k stars after 10–15 years — stars are a bad KPI; the real currencies are
  downloads (WireMock 6M/mo), container pulls, and distribution channels
  (Testcontainers modules, Helm). Growth to 500–2k stars is a realistic niche
  ceiling; distribution beats stars.

## 4. Threats
- Microcks owns multi-protocol + K8s + contract testing (CNCF gravity).
- mockd (Go, 7 protocols, MCP tools, stateful one-liner) is the closest直接
  competitor in unimock's own language — watch it.
- mockbird positions as the hosted YAML/stateful successor to stubby4j/mountebank.
- WireMock Cloud will keep absorbing "platform" buyers; OSS WireMock keeps the JVM
  install base.

## 5. Recommended unimock priorities
Lean in (already core): stateful CRUD + ID extraction, YAML-as-code, single binary,
Docker/Helm, _uni tech endpoints.
Near-term cheap wins: scoped state reset endpoints; hot reload; load-time validation
w/ precise errors; machine-readable unmatched-request reports; agents.md; Testcontainers
Go module (exists for MockServer, not for unimock — free distribution channel).
Roadmap (biggest objection removers): partial OpenAPI import (sections+scenarios from
spec; 82% of orgs are now somewhat/fully API-first, +12% YoY — import is a cheap
complement, not a repositioning), .http-request-file fixture ecosystem alignment
(go-restclient already there), outbound webhook/callback triggering from scenario
matches (WireMock has it as an extension; zero incumbents in the OSS YAML mid tier;
HTTP-native so no protocol war needed), IoT/device-simulator beachhead (validated
pattern — Nordic Semiconductor publicly uses Microcks this way — but no targeted
competitor).
Do NOT chase: gRPC/GraphQL/Kafka/AsyncAPI protocol breadth, hosted SaaS, GUI,
record/proxy as headline feature (Hoverfly's MITM complexity is the cautionary tale) —
outgunned or wrong buyers; the niche doesn't need them.
