---
# unimock-tvvr
title: 'Ponytail round 2 (optional): compact uni_handler.go 920-line dispatch'
status: draft
type: task
priority: normal
created_at: 2026-09-22T20:26:16Z
updated_at: 2026-09-22T20:26:23Z
parent: unimock-jlc4
---

uni_handler.go 48 funcs / 920 lines, switch on req.Method at :807. Candidate for table-driven dispatch or cohesive splits. Punt from round 1 — works, tested. Do only if it grows or changes.
