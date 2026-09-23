# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| latest release | ✅ |
| older releases | ❌ — please upgrade |

Unimock is a **test tool**: it is meant to run in development, CI, and test
environments. Do not expose it to untrusted networks — it executes whatever
configuration you give it and has no authentication on its admin endpoints
(`/_uni/*`).

## Reporting a vulnerability

- **Do not open a public GitHub issue.**
- Use GitHub's [private vulnerability reporting](https://github.com/bmcszk/unimock/security/advisories/new)
  for this repository.
- Include: affected version, minimal config + request reproducing the issue,
  impact assessment.

You can expect an initial response within 7 days. If the issue is confirmed,
a fix is released and credit is given unless you prefer otherwise.
