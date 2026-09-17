# Rework brief — TASK-260910-3p2rbh, revision 2 (R2: cached /health verdict)

Revision 1 was rejected with four corrections
(`TASK-260910-3p2rbh_review-verdict-rev1.md`). Corrections 1–3 are real
defects and are required as written. Correction 4 is re-decided by the
orchestrator below (the producer's semantics stand; the test must change).
Everything else passed; keep it unchanged unless a correction touches it.

## Corrections
1. **Cross-table snapshot consistency in the verifier (High).** The refresh
   reads frontier length/live size and then `_load_frontier` in separate
   snapshots; a valid concurrent append between them makes the comparison
   report corruption and latches write-disable on a healthy database. Read
   the log size, the boundary rows and the frontier under ONE consistent
   read snapshot (a single read transaction / `BEGIN` on the same connection
   with WAL snapshot isolation), or verify against the size captured first
   and compare only the frontier state as of that size. Preserve WAL append
   concurrency. Add an entry-point regression that interposes a real
   `store.append` immediately before the verifier's frontier load (the
   reviewer's reproduction) and asserts the refresh stays green.
2. **Incremental advancement must not attest a corrupted frontier (High).**
   `_append_locked` trusts the persisted frontier tails after checking only
   length/encoding, so a tampered frontier is carried into a wrong immutable
   boundary and a green verdict. Before using the frontier for an append,
   validate the anchors cheaply: recompute the new root from the frontier
   AND check the frontier against the last committed boundary row (its
   `merkle_root` must equal the root the frontier yields for `log_size`) and
   the head chain (previous entry hash = stored head); any mismatch latches
   through the common integrity-failure path (writes disabled, `/health`
   503) instead of appending. Test: corrupt length-preserving level-0 tails
   after three entries, append a fourth through the production entry —
   append refused, `/health` 503. No return to full-chain hashing per
   request.
3. **Cached verdict advances only after COMMIT (Medium).** Publish the
   `_health_verified_head/_log_size` advancement after the transaction
   commits (or restore the previous values on any failure, including a
   later import item, ledger insertion or COMMIT failure). Cover rollback
   via the real `append_imports` / idempotent append call sites (the
   reviewer's BEFORE INSERT ABORT trigger reproduction is a fine harness).
4. **Staleness is transient; corruption latches (orchestrator re-decision).**
   Keep the producer's semantics: a stale verdict (no completed refresh
   within the bound) makes `/health` 503 and refuses writes WHILE stale, and
   a subsequently completed successful refresh restores readiness (with a
   structured log event); a refresh that FAILS or finds corruption latches
   non-ready until a restart re-verifies. Update the brief-derived wording in
   README/SECURITY accordingly. The committed test must be deterministic:
   replace the sleep with an injected clock, and cover (a) the exact time
   boundary (age = bound → still fresh; age > bound → stale), (b) late
   completion of a stalled verifier restoring readiness, (c) a stalled
   verifier whose completion FAILS staying latched.

## Validation and handoff
`python -m pytest -q` with `CURATOR_CONFORMANCE_ROOT` set and `python -m mypy`
(exit codes, interpreter); update `TASK-260910-3p2rbh_results.md` with a
"Revision 2" section (per-correction file:line, transcripts) and hand off
with `task-board handoff TASK-260910-3p2rbh --role developer`. Worktree and
rules unchanged.
