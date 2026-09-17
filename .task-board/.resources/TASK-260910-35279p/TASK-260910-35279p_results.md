# TASK-260910-35279p results — memoized snapshot boundary (R2, service)

Story `STORY-260910-1py4f3`, brief `TASK-260910-35279p_brief.md`,
rules `remediation-registry-producer-rules.md`.
Worktree: `<control-root>/.temp/STORY-260910-1py4f3/worktree`
(branch `task-board/story/STORY-260910-1py4f3`, from `main` `aea81cc`).
Spec: curator-spec `dced9b8` (read-only checkout, no spec changes).

## What changed (per file)

- `src/csk_registry/store.py` — the whole fix:
  - Schema 3: new `boundaries(log_size, head, merkle_root, created_at)` and
    `merkle_frontier(level, len, tail)` tables (`store.py:43`,
    `_SCHEMA_VERSION = 3` at `store.py:70`).
  - `_append_locked` (`store.py:554`) writes the boundary row for the new
    `seq` in the same transaction as the log insert, and advances the
    durable incremental frontier (`_load_frontier` `store.py:598`,
    `_save_frontier` `store.py:624`, `_frontier_append` `store.py:1177`).
    Append Merkle cost is one pair-hash per tree level — O(log n) — and
    roots are byte-identical to the naive `_merkle_root` (`store.py:1157`).
  - `_snapshot_boundary_locked` (`store.py:912`) is now a single-row
    `boundaries` lookup plus two O(1) structural anchors (the log row at
    the boundary still carries the memoized head/timestamp; genesis row
    `seq = 1` still exists). No prefix scan, no Merkle recomputation.
    `snapshot_boundary` (`store.py:878`), `boundary_available`
    (`store.py:971`), and the R1 carried-boundary check
    (`_page_boundary_locked`, `store.py:882`) all funnel through it, so
    all are O(1) per request.
  - `_ensure_boundaries` (`store.py:636`) runs once per `Store.__init__`
    after the unchanged full-chain `integrity_errors()` walk: it
    recomputes every prefix boundary incrementally from the log and
    compares to the stored rows. A missing row is backfilled (first
    upgrade); a disagreeing row fails readiness with `StoreIntegrityError`
    like any §5 mismatch. The frontier is rebuilt, never trusted.
  - Migration: `_require_v2_schema` + `_migrate_v2_to_v3`
    (`store.py:367`, `store.py:401`) create the tables and bump markers
    atomically; legacy v0 and fresh DBs get the tables via `_SCHEMA`.
    Idempotent: reopening retries an interrupted migration; log history
    is never rewritten.
  - `_merkle_pair_hash` (`store.py:1146`) is the single choke point for
    all Merkle-tree hashing (entry-hash computation uses `hashlib`
    directly), so the operation-count test observes exactly Merkle work.
- `tests/test_registry.py` — 6 new tests (all passing, see below).
- `CHANGELOG.md` — Unreleased Security entry `R2: …` (boundary half only).
- `README.md` — operations note for the schema-3 cache tables (startup
  migration, idempotent, first-startup cost).

## How each AC line is met

Brief deliverable 1 (O(1) durable lookups):
`store.py:912` reads one `boundaries` row; rows are written per append in
the same transaction (`store.py:554`); backfill happens once at startup
(`store.py:636`). Choice justification: a `boundaries` table (not a pure
in-memory frontier) because concurrent writers are separate `Store`
instances/threads on one SQLite file — the cache must be transactional
and visible across connections, which an in-memory frontier is not. The
durable `merkle_frontier` side table keeps appends at O(log n) hashes
instead of re-scanning the prefix per append.

Brief deliverable 2 (proof stays, cache never trusted):
startup chain verification is untouched (`integrity_errors` still walks
the full chain); append still verifies against the current head
(`seq != previous_seq + 1` plus the frontier-length check in
`_append_locked`); memoized rows are revalidated against the recomputed
chain at every startup and a mismatch raises `StoreIntegrityError`
(`store.py:636`). Invalidation is exactly head advance: each new head
inserts exactly one new row; older rows are immutable history.

Brief deliverable 3 (operation-count proof, no timing):
`test_boundary_reads_perform_zero_merkle_hashes`
(`tests/test_registry.py:1285`) monkeypatches
`csk_registry.store._merkle_pair_hash` with a counting wrapper: after one
append, repeated `/v1/records` first+cursor pages, `/v1/log` first+cursor
pages, `/v1/snapshot`, and store-level `snapshot_boundary` (head and old
size), `boundary_available`, `records_page`/`log_page` with carried
boundary, `merkle_root`, `head`, `checkpoint_matches` perform **zero**
hashes (asserted `calls[0] == 0`, also after close+reopen).
`test_boundary_append_cost_is_logarithmic` (`tests/test_registry.py:1373`)
asserts one append on 64 leaves uses 1–16 hashes (observed 7: one per
level; naive would be ~64). Correctness oracle: memoized roots are
asserted equal to naive `_merkle_root` over the log leaves, so the test
proves the cache is right, not just fast. No timing assertions anywhere.

Brief deliverable 4 (concurrency green, lock discipline):
full suite green (below), including `concurrent-writers` conformance,
`test_concurrent_store_instances_serialize_writers`, idempotency,
recovery, and restore vectors. Lock discipline is unchanged: every public
method still takes `self._lock` (the shared `sqlite3` connection with
`check_same_thread=False` requires serialization), and appends still
commit log + cache rows atomically under `BEGIN IMMEDIATE`.
Nothing moved outside the write lock; what changed is lock *hold time*
(single-row SELECTs instead of full-prefix scans + hashing), which
reduces contention without weakening mutual exclusion.

Brief deliverable 5 (notes): CHANGELOG R2 entry added; README gained the
schema-3 operations note quoted above. No CI pin move needed (no new
vectors/schemas consumed; pin already `dced9b8`).

Supporting tests:
`test_merkle_frontier_matches_naive_root` (`tests/test_registry.py:1275`,
100-prefix fuzz vs naive),
`test_boundary_cache_disagreement_fails_startup_but_missing_row_backfills`
(`tests/test_registry.py:1399`, tampered row → `StoreIntegrityError` at
open; deleted row + frontier → backfilled, reads correct),
`test_v2_database_migrates_and_backfills_boundaries`
(`tests/test_registry.py:1451`, downgraded v2 file migrates to 3 with
correct roots),
`test_frontier_tampering_rebuilds_and_next_append_stays_correct`
(`tests/test_registry.py:1501`).

## Validation transcripts (exit codes real, `set -o pipefail`)

Shell: `/bin/sh` via run harness, worktree root.
Interpreter: `/tmp/csk-venv/bin/python`, Python 3.14.6
(venv outside the tree per campaign rules; CI also covers 3.11 — not run
locally).

- `export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1; /tmp/csk-venv/bin/python -m pytest -q`
  → `139 passed, 2 warnings in 13.79s`, exit `0`
  (baseline before the change: `133 passed`; +6 new tests).
- `/tmp/csk-venv/bin/python -m mypy` (strict, per `pyproject.toml`)
  → `Success: no issues found in 13 source files`, exit `0`.
- Hosted gate (`scripts/remote-gate.sh`, GitHub CI matrix) NOT run by me
  per the campaign rules — the runtime runs it once at handoff.

## Deliberately out of scope

`/health` cached verdict + background verifier (`TASK-260910-3p2rbh`);
protocol/profile changes; key rotation. `integrity_errors()` (and hence
`/health`) intentionally does NOT revalidate the Merkle cache per call —
that stays a startup-only cost so this leaf does not make health probes
slower before the sibling lands.

## Design trade-off worth knowing (not a spec gap)

Per-request full-chain verification is gone by design (that was the R2
finding). Live reads now verify two O(1) anchors — boundary head row +
genesis existence — so behind-the-back prefix pruning is still refused
immediately as `invalid_cursor` (existing pruned-prefix test stays
green), while interior tampering is caught at the next startup instead of
the next request. The memoized-row-vs-log comparison at startup is the
fail-closed backstop. No spec text changes needed; no spec gaps found.
