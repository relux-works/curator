# TASK-260910-3p2rbh results — cached /health verdict with background verifier (R2, service)

Story `STORY-260910-1py4f3` (boundary-verification-performance,
`EPIC-260910-16qce1`), wave 3, final leaf. Builds on the checkpointed
`TASK-260910-35279p` commit (`693df37`, schema-3 boundaries/frontier tables).
Rules: `remediation-registry-producer-rules.md`. Spec: curator-spec `dced9b8`
read-only checkout, no spec changes. Worktree:
`<control-root>/.temp/STORY-260910-1py4f3/worktree`
(branch `task-board/story/STORY-260910-1py4f3`).

## What changed (per file)

- `src/csk_registry/store.py` — the verdict, the verifier, the fail-closed gate:
  - `HealthVerdict` (`store.py:148`): frozen `{ready, verified_head,
    verified_log_size, verified_at, error}`. Read-time staleness only; no
    hashing or I/O to read (`health_verdict()`, `store.py:1230`;
    staleness predicate `store.py:1225`).
  - `Store.__init__` takes keyword-only `health_verify_interval`
    (validated positive finite, `store.py:188`-202; default
    `DEFAULT_HEALTH_VERIFY_INTERVAL_SECONDS = 300.0`, `store.py:78`;
    staleness bound `HEALTH_STALENESS_MULTIPLIER = 2.0`, `store.py:82`).
    Fail-closed defaults first; the first verdict is published from the
    startup §5 result via one O(1) `head()` read (`store.py:268`-278), so a
    fresh service is ready immediately after verification passes.
  - `_require_writes_allowed()` (`store.py:628`): O(1) gate called at the
    entry of `append`, `append_idempotent`, `append_imports`. Latched
    failure or staleness raises `StoreIntegrityError` — the same exception
    type as the startup §5 mismatch, so the app layer maps it to `503`
    exactly like a startup failure.
  - `_append_locked` (`store.py:697`): after the existing contiguity /
    prev-linkage / frontier-length checks pass, the verified head/size
    advance incrementally. `verified_at` (the staleness clock) is
    deliberately untouched, so interior tampering still needs a full pass.
  - `_integrity_errors_on(conn)` (`store.py:1093`): the full
    chain-plus-ledgers walk extracted verbatim (same error strings) into a
    connection-parameterized choke point returning `(errors, leaves)`.
    Startup (`integrity_errors`, `store.py:1088`) uses it on the store
    connection under lock with the 30 s deadline; the refresh uses it on a
    short-lived `PRAGMA query_only` connection holding no store lock. The
    only logic delta is a live single-row fallback for idempotency seqs
    missing from the walked prefix (identical outcomes at startup, where
    the prefix is everything; race-safe at refresh).
  - `_verify_boundary_cache_on` (`store.py:1263`) + pure
    `_boundary_cache_errors` (`store.py:1678`): the refresh also compares
    the memoized `boundaries` rows and `merkle_frontier` against the
    recomputed prefix (`_recompute_prefix`, `store.py:1660`, now shared
    with `_ensure_boundaries`). Stored rows strictly between the walked
    size and the live head are ignored as concurrent appends; rows beyond
    the live head, missing/disagreeing rows, and frontier length/tail
    mismatches are corruption. Frontier length vs live head is read in one
    statement (race-free); tails are skipped only while appends are
    landing, length still enforced.
  - `refresh_health_verdict()` (`store.py:1309`): one total pass (returns a
    verdict, never raises). Walk holds no store lock; only publication
    takes `self._lock` briefly. No operation deadline on the walk — its
    duration scales with log size and the staleness bound is the backstop.
    ANY failed pass (corruption found or the walk raising) latches
    `_integrity_failed`: non-ready + writes disabled until a restart
    re-verifies. A clean pass publishes the live head (walked prefix +
    transitively-trusted incremental appends). Every outcome is logged to
    the `csk_registry.audit` logger (`store.py:1421`: `health_refresh`
    with `result`/`duration_ms`/`log_size`, plus `error_count`/`error` on
    failure).
  - Verifier thread (`store.py:1440`-1490): idempotent
    start/stop/`health_verifier_running()`; daemon looping on the
    interval; lifecycle lock separate from `self._lock`; bounded 10 s
    join; `close()` (`store.py:498`) stops the verifier before closing
    the connection (never under the store lock).
- `src/csk_registry/app.py` — `GET /health` serves the cached verdict only
  (`app.py:201`; success envelope unchanged per frozen
  `health-response-v1`; failure `503 not_ready`); lifespan starts/stops
  the verifier (`app.py:98`); `app_from_env` passes
  `CURATOR_SKILL_REGISTRY_HEALTH_VERIFY_INTERVAL` through
  `_positive_float_env` (`app.py:648`, `app.py:686`).
- `src/csk_registry/__init__.py` — `HEALTH_VERIFY_INTERVAL_ENV` constant
  (`__init__.py:10`).
- `src/csk_registry/cli.py` — `serve --health-verify-interval SECONDS`
  (`cli.py:469`; default `None` = env/default; validated positive finite
  before boot, `cli.py:343`; exported to the env var, `cli.py:351`).
- `tests/test_registry.py` — 4 helpers + 11 tests (`test_registry.py:1501`-2080).
- `CHANGELOG.md` — R2 entry completed (health half replaces the
  "follows separately" tail). `README.md` — operations note (interval,
  sizing via `verify-chain`, staleness bound, what non-ready means, audit
  event). `SECURITY.md` — fail-closed health paragraph. `compose.yaml` —
  comment that the 30 s probe is a cheap cached read (no functional change).

## How each AC line is met

Brief: "Health no longer runs the full chain per probe; corruption still
flips readiness."

1. Cached verdict + background verifier, zero per-probe hash work, first
   verdict = startup verification:
   `app.py:201` reads `store.health_verdict()` (`store.py:1230`, lock +
   clock only). `test_health_probes_perform_zero_hash_work`
   (`test_registry.py:1564`) counts `sha256` (store-namespace swap),
   store `canonical_bytes`, and `_merkle_pair_hash` across 5 probes, then
   2 appends, then 5 more probes: all three counters are **0** on every
   probe batch. `test_health_first_verdict_is_startup_verification`
   (`test_registry.py:1605`) reopens a pre-populated DB and asserts
   `/health` is `{"status": "ok"}` with no refresh driven and no thread
   started. Call-site audit: `integrity_errors()` now runs only at
   `Store.__init__`/migration/`verify_chain` (`store.py:249,304,1086`);
   no request handler walks the chain.
2. Fail closed (`503`, writes disabled, latch, restart recovery):
   - Corruption → next refresh flips: `test_health_corruption_detected_at_next_refresh_disables_writes`
     (`test_registry.py:1634`) tampers `log.entry_hash` behind the back,
     asserts `/health` stays 200 pre-refresh (cached — negative control),
     drives `refresh_health_verdict()` explicitly (no sleeps), then
     asserts `503 not_ready`, POST → `503 storage_unavailable`,
     `store.append` raises (`restart to re-verify`), and the audit log
     holds `integrity_failed` then `integrity_failed_latched` with
     `duration_ms`/`log_size`/`error_count`.
   - Stale → `503` + writes refused, recovers on refresh (not latched):
     `test_health_stale_verifier_fails_closed_and_refresh_recovers`
     (`test_registry.py:1692`, interval 0.05 s, one-sided 0.2 s sleep past
     the 0.1 s bound).
   - Latch + restart after repair → ready:
     `test_health_restart_after_repair_returns_ready`
     (`test_registry.py:1724`): repair + refresh without restart stays
     non-ready; reopening the repaired DB is ready and appends (seq 3).
   - Incremental head advance without touching the staleness clock:
     `test_health_append_advances_verified_head_without_full_refresh`
     (`test_registry.py:1758`, `verified_at` unchanged across append).
   - Recovery vectors + §9 behavior green: full suite below (all
     `recovery_cases`, restore, transport, and pre-existing health/boundary
     tests pass unmodified).
3. Full verifier class coverage (what the pass checks):
   `test_health_refresh_detects_boundary_cache_tampering`
   (`test_registry.py:1831`) tampers only `boundaries.merkle_root` — live
   O(1) reads keep serving it (asserted) while the refresh flags
   "boundary cache"; `test_health_refresh_detects_frontier_tampering`
   (`test_registry.py:1858`) flips a length-preserving frontier tail →
   "frontier" flagged. `test_boundary_cache_comparison_branches`
   (`test_registry.py:1888`) unit-tests the pure comparator: exact match,
   missing/disagreeing row, concurrent-append row ignored + tails skipped,
   row beyond live head, invalid size, length mismatch, tail mismatch,
   empty/empty.
4. Thread + wiring: `test_health_verifier_lifecycle_and_lifespan`
   (`test_registry.py:1777`): idempotent start/stop, a 0.05 s interval
   fires a real background pass within a bounded 10 s poll, and the
   serving lifespan starts/stops the thread. Flag/env parsing:
   `test_serve_health_verify_interval_flag_and_env`
   (`test_registry.py:1987`, invalid values rejected before boot).
5. Transcripts / docs: below; CHANGELOG R2 completed; README/SECURITY/
   compose operations note added.

## Validation transcripts (exit codes real)

Shell `/bin/sh` (via harness), worktree root.
Interpreter `/tmp/csk-venv/bin/python`, Python 3.14.6 (venv outside the
tree; CI also covers 3.11 — not run locally).
Conformance root exported for pytest per campaign rules.

- `export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1; /tmp/csk-venv/bin/python -m pytest -q`
  → `150 passed, 2 warnings in 5.94s`, exit `0`
  (baseline at checkpoint: `139 passed`; +11 new tests; the 2 warnings are
  the pre-existing starlette/httpx deprecation notices, also present at
  baseline).
- `/tmp/csk-venv/bin/python -m mypy` (strict, per `pyproject.toml`)
  → `Success: no issues found in 13 source files`, exit `0`.
- `/tmp/csk-venv/bin/python -m build --outdir /tmp/3p2rbh-dist`
  → sdist + wheel built, exit `0` (packaging smoke; output outside the tree).
- No configured linter exists in the repo (no ruff/flake8 config; CI gates
  are pytest + mypy + build + docker) — "lint clean" is satisfied via the
  mypy-strict gate above, exit `0`.
- Hosted gate (`scripts/remote-gate.sh`, GitHub CI matrix) NOT run by me
  per the campaign rules — the runtime runs it once at handoff.
- CI pin: already `dced9b8...` in `.github/workflows/ci.yml`; no new
  vectors/schemas consumed, no pin move needed.

## Lock discipline (as the brief requires it stated)

The verifier walk holds **no** store lock: it opens its own short-lived
read-only (`PRAGMA query_only`) SQLite connection per pass, so a
minutes-long walk on a large log never blocks appends (WAL readers don't
block writers). `self._lock` is taken only to publish the verdict (field
writes + one O(1) boundary read). Thread start/stop use a separate
lifecycle lock; `close()` stops the verifier before closing the
connection. Request paths are unchanged: `self._lock` everywhere, plus an
O(1) verdict gate on the three write entries.

## Deliberately out of scope

Protocol/profile edits, key rotation, R3/P2 checkpoint gate
(`TASK-260910-1ny7yl`); the manager client; tags/releases; Docker/deploy
changes beyond the compose comment. No Migrations beyond what 35279p
landed (schema stays 3). DB-file replacement without restart is still
undetected by design (no high-water comparison — that is the R3/P2 story).

## Design notes / trade-offs

- Any failed pass latches (including a walk that raises, e.g.
  `SQLITE_BUSY` past the 5 s timeout): fail-closed simplicity — restarts
  are cheap and re-verify. No deadline on the background walk (duration
  scales with log size); the staleness bound is the backstop. Operators
  must size the interval above the full-walk time (README documents
  `time … verify-chain` as the proxy); otherwise a large log would sit
  permanently stale — that is the honest, documented failure mode, and it
  fails to non-ready, never to false-ready.
- No spec gaps found; no spec text changes needed. The frozen
  `health-response-v1` success envelope is byte-unchanged; only the
  failure detail collapsed from two messages to one (`not_ready`,
  "registry durable state failed integrity verification") since probe-time
  failures no longer exist as a category.
- Skills: none of the role-catalogued skills (swiftui, core-data,
  go-testing-tools, architecture-diagrams) match this Python-service
  scope, so no bodies were bulk-read; `bundled:durable-test-collateral`
  was loaded before the first edit and its maintained-test rule is
  satisfied by the 11 committed tests above.

---

# Revision 2 (rework per `TASK-260910-3p2rbh_review-verdict-rev1.md` + orchestrator re-decision `TASK-260910-3p2rbh_rework-rev2.md`)

Revision 1 was rejected with four corrections. Corrections 1–3 are fixed as
written; correction 4 keeps the producer's transient-staleness semantics per
the orchestrator re-decision (stale recovers on the next successful pass,
failed/corrupt latches until restart) with deterministic clock-injected
tests. Everything else from revision 1 is unchanged unless a correction
touched it.

## Per-correction file:line

1. **Cross-table snapshot consistency (High).**
   `src/csk_registry/store.py:1492` — `refresh_health_verdict()` now opens
   `BEGIN` (DEFERRED) on its `isolation_level=None` query-only connection
   (`store.py:1484`) and holds one WAL read snapshot across
   `_integrity_errors_on()` + `_verify_boundary_cache_on()`, ending with
   `ROLLBACK` (`store.py:1498`-1502). Docstrings updated:
   `refresh_health_verdict()` snapshot discipline (`store.py:1437`-1465),
   `_verify_boundary_cache_on()` (`store.py:1384`-1404),
   `_integrity_errors_on()` (`store.py:1212`-1230). WAL readers never block
   writers; the pass only pins the snapshot. Regression:
   `test_health_refresh_sees_concurrent_append_atomically`
   (`tests/test_registry.py:1801`) interposes a real `store.append`
   immediately before the verifier's `_load_frontier` and asserts the
   refresh stays green (`verified_log_size == 2`, `/health` 200,
   `integrity_errors() == []`).
2. **Incremental anchors must not attest a corrupted frontier (High).**
   `src/csk_registry/store.py:744` — new
   `_validate_append_anchors_locked()` (O(1) rows + one hash per tree
   level, no chain walk): frontier length == `previous_seq`; last
   `boundaries` row exists with `head == prev_hash`; level-0 tails equal
   the last 1–2 committed leaves; each upper level's last node equals the
   hash of the lower tails; top root equals the boundary `merkle_root`.
   Any mismatch latches via `_fail_append_locked()` (`store.py:667`) →
   `_latch_integrity_failure_locked()` (`store.py:645`: sets
   `_integrity_failed`, `ready=False`, audit `health_refresh`
   `integrity_failed`) and raises `StoreIntegrityError` instead of
   appending. `_append_locked()` (`store.py:684`) validates before the log
   INSERT and latches frontier-encoding, contiguity, `_frontier_append`,
   and boundary-INSERT failures too; `ValueError` for malformed records
   does not latch. Test:
   `test_health_append_refuses_corrupted_frontier_and_latches`
   (`tests/test_registry.py:1829`): length-preserving level-0 tail swap
   after three entries → fourth `store.append` raises (`match="frontier"`),
   `/health` 503 `not_ready`, durable head still 3 with the genuine root
   (`_merkle_root` over the committed log), refresh stays latched.
   Bound (stated): second-last tails of odd-length upper levels (length ≥
   3) are not re-derivable from O(log n) anchors; such an exotic swap is
   still caught at the next full refresh (window ≤ interval) — the
   committed boundary comparison recomputes every prefix from the log.
3. **Cached verdict advances only after COMMIT (Medium).**
   `_append_locked()` no longer touches `_health_*` on success; publication
   moved to the callers after the write transaction commits:
   `append()` (`store.py:672`-682), `append_idempotent()`
   (`store.py:923`-979, publish at `:975`-977), `append_imports()`
   (`store.py:1065`-1094, publish at `:1088`-1093 to the last durable
   entry). Any failure — later import item, ledger INSERT, or COMMIT
   itself — skips publication, leaving the previous durable verdict.
   Tests: `test_health_rollback_does_not_advance_cached_verdict`
   (`tests/test_registry.py:1874`, BEFORE INSERT ABORT trigger on
   `imported_records` via real `append_imports`) and
   `test_health_idempotent_rollback_does_not_advance_cached_verdict`
   (`tests/test_registry.py:1898`, trigger on `idempotency` via real
   `append_idempotent`): both assert durable head 0, verdict size 0,
   `/health` 200.
4. **Staleness transient, corruption latches (orchestrator re-decision).**
   No semantic change to `store.py` staleness (`_health_stale_locked()`,
   `health_verdict()`, `refresh_health_verdict()` success path): stale
   (age > 2 × interval) → 503 + writes refused while stale; next
   successful pass restores readiness with a `health_refresh ok` event;
   failed/corrupt pass latches until restart. Wording updated:
   `README.md:183`-190 (transient vs latch), `SECURITY.md:32`-41,
   `CHANGELOG.md:51`-56. Tests deterministic (no sleeps):
   `_ManualStoreClock` (`tests/test_registry.py:1692`) injected via
   `monkeypatch.setattr(store_module, "time", clock)`;
   `test_health_stale_verifier_fails_closed_and_refresh_recovers`
   (`tests/test_registry.py:1702`, interval 10, bound 20) covers (a) exact
   boundary (age 20.0 → 200 + append allowed; 20.001 → 503 + `stale`
   refusal) and (b) late successful completion restoring readiness and
   writes, then fresh-at-20.0 / stale-at-20.001 on the renewed clock;
   `test_health_stalled_verifier_failure_stays_latched_until_restart`
   (`tests/test_registry.py:1755`) covers (c): stall + corrupt → failed
   refresh latches; fresh clock, write refusal (`restart to re-verify`),
   repair + refresh without restart all stay 503; reopen after repair →
   ready.

## Revision 2 validation transcripts (exit codes real)

Shell `/bin/zsh` (harness `bash` tool), worktree root.
Interpreter `/tmp/csk-venv/bin/python`, Python 3.14.6 (venv outside the
tree; CI also covers 3.11 — not run locally).
Conformance root exported for pytest per campaign rules.
`set -o pipefail` semantics: each gate ran as a standalone process, no
`tee`/pipe chain.

- `export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1; /tmp/csk-venv/bin/python -m pytest -q`
  → `155 passed, 2 warnings in 11.98s`, exit `0`
  (rev1 baseline: `150 passed`; +5 new: concurrent-snapshot, frontier-latch,
  2× rollback, stalled-failure-latch; stale test rewritten deterministically
  in place. The 2 warnings are the pre-existing starlette/httpx deprecation
  notices.)
- `/tmp/csk-venv/bin/python -m mypy` (strict, per `pyproject.toml`)
  → `Success: no issues found in 13 source files`, exit `0`.
- Revision 1 reviewer reproductions re-run against the rev2 tree (disposable
  copy of `TASK-260910-3p2rbh_review-attacks-rev1.py`, then removed):
  `test_review_concurrent_append_between_verifier_reads` PASS,
  `test_review_append_does_not_attest_corrupted_frontier` PASS,
  `test_review_rollback_does_not_advance_cached_head` PASS,
  `test_review_frontier_corruption_creates_wrong_boundary` now correctly
  FAILS (it asserts the old buggy attestation: append succeeding with a
  wrong root while health stays 200 — rev2 refuses the append).
  `test_review_staleness_must_latch_until_restart` obsolete per the
  orchestrator re-decision (staleness is transient by design).
- No configured linter (no ruff/flake8 config; CI gates are pytest + mypy +
  build + docker) — "lint clean" via the mypy-strict gate above, exit `0`.
- Hosted gate (`scripts/remote-gate.sh`) NOT run by me per campaign rules —
  the runtime runs it once at handoff.
- CI pin: already `dced9b8...` in `.github/workflows/ci.yml`; no new
  vectors/schemas consumed, no pin move needed.

## Revision 2 lock/operation-count notes

- Refresh walk still holds no store lock; the added `BEGIN`/`ROLLBACK` is on
  the pass-private connection only. Publication still takes `self._lock`
  briefly. Appends still serialize on `self._lock` + `BEGIN IMMEDIATE`.
- Zero-hash probes unchanged: `test_health_probes_perform_zero_hash_work`
  still passes unmodified (5 probes + 2 appends + 5 probes, all counters 0
  on probe batches). Append Merkle cost stays logarithmic:
  `test_boundary_append_cost_is_logarithmic` passes unmodified (65th
  append within 1–16 hashes; anchor validation adds ~1 hash per level).
- Audit: append-time latches emit `health_refresh` `integrity_failed`
  (duration 0, current verified size); refresh outcomes unchanged
  (`ok` / `integrity_failed` / `integrity_failed_latched` with
  `duration_ms`/`log_size`/`error_count`).

## Out of scope (unchanged)

Protocol/profile edits, key rotation, R3/P2 checkpoint gate
(`TASK-260910-1ny7yl`); manager client; tags/releases; Docker/deploy beyond
the compose comment. Schema stays 3. DB-file replacement without restart
still undetected by design (R3/P2 story).
