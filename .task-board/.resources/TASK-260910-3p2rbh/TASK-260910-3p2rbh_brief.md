# Brief — TASK-260910-3p2rbh: cached /health verdict with a background verifier (R2, service)

Story `STORY-260910-1py4f3` (boundary-verification-performance,
`EPIC-260910-16qce1`), wave 3; second and FINAL leaf. The first leaf
`TASK-260910-35279p` (accepted, checkpointed on the Story branch) made
boundary lookups O(1) with durable `boundaries` / `merkle_frontier` tables
(schema 3) rebuilt and verified at startup. Read `TASK-260910-35279p_results.md`
and `src/csk_registry/store.py` (`integrity_errors()`, the schema-3 tables,
the startup verification) before coding; build on it. Rules:
`remediation-registry-producer-rules.md` (attached). Role: developer.
Worktree: the managed Story worktree
`<control-root>/.temp/STORY-260910-1py4f3/worktree` (carrying the checkpointed
35279p commit).

## Finding (read it first)
`docs/security-audit-2026-09.md` R2 (Medium), the `/health` half:
`integrity_errors()` re-verifies the entire chain and canonical bytes on
EVERY `/health` call; compose probes it every 30 s; large logs degrade into
self-DoS. Profile §9: `/health` returns success only when the store is
readable, the verified head matches the immutable snapshot boundary, and
required dependencies are available; §5: a mismatch in authoritative state
fails readiness and disables writes.

## Settled decisions
- `/health` serves a **cached integrity verdict**: `{ready, verified_head,
  verified_log_size, verified_at, error}` held in the store/app and refreshed
  by a **background full verifier** (thread) on a bounded interval (working
  name `--health-verify-interval`, default e.g. 300 s, documented) and after
  every append the cheap incremental check (new entry against head, frontier
  consistency) updates `verified_head`; a full chain re-verification never
  runs on the request path.
- **Fail closed**: a failed or stale refresh makes `/health` non-ready
  (`503` with the protocol error envelope): stale = the verifier has not
  completed within `2 × interval` (or the configured bound) — a hung
  verifier must not leave a green cached verdict forever; a verifier that
  finds corruption flips readiness AND disables writes exactly like the
  startup §5 mismatch (same code path/flag), and stays non-ready until a
  restart re-verifies. Corruption introduced between refreshes (e.g. a test
  mutating the log file) is detected at the next refresh — the test drives
  the refresh explicitly (no sleeps).
- The first verdict is the startup verification result (so a freshly
  started service is ready immediately after §5 verification passes).
- The verifier holds the read side of the store discipline established by
  35279p (say what it locks and for how long; it must not block appends for
  the whole walk — verify in bounded chunks or under a snapshot of the log
  size).
- Observability: the audit/structured log records each refresh outcome
  (duration, log_size, result); the frozen `health-response-v1` schema is
  not changed.

## Deliverable
1. The cached verdict + background verifier + fail-closed staleness in
   `app.py`/`store.py`/`cli.py` (flag, default, `compose.yaml`/README note).
2. Tests: `/health` performs zero full-chain hash work per probe (operation
   counter as in 35279p); corruption injected after startup → next refresh
   flips `/health` to `503` and writes are refused; stale verifier → `503`;
   restart after repair → ready; existing `recovery_cases` and the profile
   §9 conformance still green.
3. `CHANGELOG.md` R2 entry completed; README/SECURITY operations note
   (interval, staleness bound, what a non-ready health means).

## Out of scope
Protocol/profile edits, key rotation, R3/P2 checkpoint gate
(`TASK-260910-1ny7yl`, separate story).

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-3p2rbh_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`, operation-count evidence), then
`task-board handoff TASK-260910-3p2rbh --role developer`.
