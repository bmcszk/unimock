---
# unimock-3zla
title: 'Ponytail round 2 (optional): split 20 bool-flag params into explicit methods'
status: draft
type: task
priority: normal
created_at: 2026-09-22T20:26:16Z
updated_at: 2026-09-22T20:26:22Z
parent: unimock-jlc4
---

UniService/UniStorage take isStrictPath bool through 20 call sites. Go-idiomatic alternative: explicit GetStrict/GetFlexible-style methods at the service boundary. Punt from round 1 — bigger blast radius, zero functional gain. Do only if touching this code anyway.
