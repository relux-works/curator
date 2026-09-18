# Brief — TASK-260910-2c7s0u: deployment documentation for proxy rate-limit bucketing and body-before-auth bounds (R4, service docs)

Story `STORY-260910-2xe3n2` (registry-robustness-hardening, `EPIC-260910-16qce1`),
wave 4; fourth and FINAL leaf (after R5, R8, R7 — all checkpointed on the
Story branch; build on them). Rules: `remediation-registry-producer-rules.md`
(attached). Role: doc-writer (docs task; no code change unless a doc
example must be corrected).
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-2xe3n2/worktree`.

## Finding (read it first)
`docs/security-audit-2026-09.md` R4 (Low): `submit` streams up to 16 MiB
(15 s deadline) before token verification; the concurrency semaphore
(128, 0.1 s acquire) can be occupied by unauthenticated slow streams
(`503 overloaded`) — bounded; but behind a reverse proxy without trusted
forwarded headers the network limiter keys on the proxy IP, so one noisy
client throttles everyone.

## Deliverable
1. `README.md` deployment section (or `docs/DEPLOYING.md` if the README
   already delegates deployment there — follow the repository's existing
   structure; do not create a second place): (a) the rate-limit model —
   which limiter keys on what, that a reverse proxy without trusted
   forwarded headers collapses all clients into one network bucket, and
   the recommended configuration (terminate TLS at the proxy, enable
   forwarded-header trust ONLY for the proxy's addresses — cite the exact
   setting/flag/env the service exposes today, with its current default,
   quoting `README.md:160` "forwarded headers from other sources are
   ignored" and the code path), with a worked nginx/Caddy/Traefik snippet
   for the recommended one; (b) the body-before-auth squatting bounds
   (16 MiB, 15 s, 128 slots, 0.1 s acquire; what an attacker can and
   cannot do with them; `503 overloaded` semantics) and the proxy-side
   mitigations (request body size limits, connection limits per client);
   (c) a short operator checklist.
2. `SECURITY.md`: one paragraph pointing to that section from the threat
   model; `CHANGELOG.md` Unreleased entry "R4: …".
3. If the documented limits or settings do not match the code, report the
   mismatch in the results and document the CODE's actual behaviour (never
   change code in a docs task; a mismatch is a finding for the orchestrator).

## Out of scope
Implementing forwarded-header trust changes, limiter changes, spec edits.

## Checklist and handoff
Tick the checklist items; attach `TASK-260910-2c7s0u_results.md` (file:line,
link check of the new docs, `python -m pytest -q` unchanged-green transcript),
then `task-board handoff TASK-260910-2c7s0u --role doc-writer`.
