# TASK-260910-2g5v17 results — service-import-high-water (P4)

Story `STORY-260910-stz5f0`, worktree
`.temp/STORY-260910-stz5f0/worktree` on branch
`task-board/story/STORY-260910-stz5f0` (base `c7ef32c`).
Interpreter: ` /tmp/csk-venv/bin/python` 3.14.6 (system `python3` 3.14.6).
Conformance root:
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`.
CI protocol-suite pin left untouched (`47c3c8c`, already newer than `dced9b8`;
this task consumes no new vectors or schemas).

## Revision 2 (current) — F1: authoritative comparison under writer serialization

Rework brief `TASK-260910-2g5v17_rework-rev2.md`: round-1 review passed every
serial scenario, both mutants, atomicity, storage protection and docs; one
medium finding (F1) remained — the pre-transaction high-water read plus the
generic `ValueError("upstream high-water advanced during import")` in the
store guard lost the closed diagnostics, the override policy and the audit
event under three legal competing-writer interleavings. Everything else is
kept as in revision 1 (docs byte-identical: no README/SECURITY/CHANGELOG
change in revision 2).

### Revision-2 delta, per file

- `src/csk_registry/store.py:188,190` — the closed diagnostics
  `import_upstream_rollback` / `import_upstream_inconsistent` now live in the
  store module (single spelling owner; `bundle.py` re-exports them).
- `src/csk_registry/store.py:193` — `UpstreamHighWaterConflict(ValueError)`
  carries `diagnostic`, `key_id`, the authoritative `persisted` boundary and
  the `offered` boundary; its message names the `key_id` and both boundaries.
- `src/csk_registry/store.py:235` — `UpstreamImportResult` (`imported`,
  authoritative before-image `persisted`, `outcome` in
  `advanced`/`noop`/`accepted_older`).
- `src/csk_registry/store.py:251` — `_upstream_offer_diagnostic`, the single
  §5 comparison (below → rollback; equal+different → inconsistent; otherwise
  acceptable).
- `src/csk_registry/store.py:1360` — new `append_upstream_import(...)`: the
  authoritative comparison runs inside the serialized write transaction
  (`BEGIN IMMEDIATE`) against the latest committed high-water, with the full
  contract — rollback refusal (or warn-import without lowering under
  `accept_older_upstream`), never-overridable inconsistent refusal, identical
  no-op persisting nothing, advance in the same transaction as the records;
  any refusal or mid-batch failure rolls everything back.
- `src/csk_registry/store.py:1283` — legacy `append_imports` guard now raises
  `UpstreamHighWaterConflict` with the proper closed diagnostic instead of
  the generic `ValueError` (no production caller passes a high-water row any
  more; kept as closed defense-in-depth for direct callers).
- `src/csk_registry/bundle.py:13-17,31` — imports the diagnostics/conflict
  from the store and re-exports them (`cli.py`'s
  `from .bundle import IMPORT_UPSTREAM_ROLLBACK` keeps working).
- `src/csk_registry/bundle.py:101,160-217` — `import_bundle` drops the
  pre-transaction read entirely: verify → countersign → one
  `append_upstream_import` call; a conflict is logged as the refusal audit
  event with the authoritative `persisted` boundary and re-raised as the
  typed `UpstreamRollbackError`/`UpstreamInconsistentError`; success is
  logged from the authoritative `UpstreamImportResult` (`ok`/`noop`,
  `warning`+`accepted_older` for the override path).
- `src/csk_registry/bundle.py:43,66` — typed errors unchanged (messages still
  name `key_id` + both boundaries; still `ValueError`s, so the CLI/exit-code
  contract is unchanged).
- `tests/test_registry.py:2966` — serial identical-reimport test hardened:
  the re-import runs under a frozen distinct `utc_now`, so row equality
  proves no rewrite (previously a same-second rewrite was invisible because
  `utc_now` has one-second resolution — found via mutant M2 below).
- `tests/test_registry.py:3268,3297,3320` — `_upstream_triple` (v1/v2/v3
  history), `_pause_outer_import_at_write` deterministic competing-writer
  harness (real second `Store` + real `import_bundle`, competitor commits
  fully before the outer write runs; no threads, no sleeps), `_import_events`
  audit parser.
- `tests/test_registry.py:3328` — race 1: persisted v1, outer v2,
  competitor commits v3 first → `import_upstream_rollback` naming `key_id`,
  persisted v3 head and offered v2 head, refusal audit event
  (`persisted.version==3`, `offered.version==2`), outer persisted nothing
  (high-water row and `(log_size, head)` equal the competitor's committed
  state).
- `tests/test_registry.py:3373` — race 2: same schedule with
  `accept_older_upstream=True` → returns 0 (v2's records dedup against the
  committed v3), high-water stays v3, `ok` audit event with
  `warning=import_upstream_rollback` + `accepted_older=true` and the
  authoritative v3/v2 boundaries.
- `tests/test_registry.py:3421` — race 3 (parametrized `accept_older`
  False/True): competing first imports with different v1 bodies →
  `import_upstream_inconsistent` naming `key_id` and both heads, refusal
  audit event, competitor's committed v1 stands byte-identical.
- `tests/test_registry.py:3488` — race 4: outer v2, competitor commits
  identical v2 first → returns 0, high-water row (including `updated_at`,
  via distinct frozen import timestamps) and log equal the competitor's
  committed state, `noop` audit event with persisted/offered v2.

Serial AC mapping (fresh line numbers; production sites moved): first import
`tests:2880` → `bundle.py:178` + `store.py:1360`; rollback refused `tests:2898`
→ `bundle.py:184-188` + `store.py:1393-1404`; inconsistent `tests:2925` →
same; identical no-op `tests:2966` → `store.py:1405-1407` + `bundle.py:201`;
newer advances `tests:2993` → `store.py:1423-1436`; flag path `tests:3013,3149`
→ `store.py:1395-1400` + `bundle.py:189`; failed import untouched
`tests:3058,3086` → verify-before-compare + single-transaction
`store.py:1385-1442`; per-key `tests:3112`; migration `tests:3199`; backup
`tests:3234`. Backup/verify-backup choice, out-of-scope list and spec-gap
statement from revision 1 are unchanged.

### Revision-2 validation transcripts

Shell bash, cwd the Story worktree, no pipes on gate commands (exit codes are
the commands' own). Same venv (`/tmp/csk-venv/bin/python`, 3.14.6, editable
install of this worktree). Conformance root corrected to the exact CI pin per
the corrected rules: detached worktree `/tmp/spec-47c3c8c` at
`47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`
(`CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1`).

Full suite (second run; see the flake note below for the first):

```
$ CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q -p no:cacheprovider
........................................................................ [ 39%]
........................................................................ [ 78%]
.......................................                                  [100%]
183 passed, 2 warnings in 38.73s
PYTEST_EXIT:0
```

(183 = 178 revision-1 tests + 5 new race tests. The 2 warnings are the
pre-existing fastapi/starlette dependency deprecations.)

```
$ /tmp/csk-venv/bin/python -m mypy
Success: no issues found in 14 source files
MYPY_EXIT:0
```

`git diff --check` exit 0. No linter is configured (CI gates are pytest +
mypy strict + build + docker); the hosted `scripts/remote-gate.sh` gate is
not run by the producer (runtime runs it at handoff).

Flake note (reported honestly, not hidden): the first full-suite run exited 1
with 182 passed and 1 failed —
`test_concurrent_store_instances_serialize_writers` (`StoreIntegrityError:
log changed during startup boundary verification` from `_ensure_boundaries`
after its 4 retries). That test is a pre-existing 8-thread open/append/close
stress test on a code path this change does not touch (`Store.__init__` /
`_ensure_boundaries` / `append` are byte-identical to base `c7ef32c`). It
passes 5/5 in isolation on this tree and 6/6 in isolation on a pristine base
checkout in /tmp (module path verified as the base copy), and the immediate
full-suite re-run above is green (183 passed, exit 0). Assessment: load flake
under a busy box, not a regression; left for the round-2 reviewer to re-run
independently.

Mutants (planted in disposable /tmp copies; worktree untouched; real entry
points driven):

- M1 (F1 regression: serialized comparison reverted to generic `ValueError`,
  override branch removed): the 4 refusal/override race tests fail, exit 1;
  the race no-op test passes (it does not exercise the refusal path, as
  designed). Caught.
- M2 (narrowing: identical boundary reports `noop` but still rewrites the
  high-water row): initially escaped both noop tests because `utc_now` has
  one-second resolution and same-second imports stamp identical `updated_at`;
  after freezing distinct import timestamps in `test_identical_reimport_is_noop`
  and `test_concurrent_identical_import_is_noop`, both fail on M2, exit 1.
  Caught.

---

## Revision 1 (superseded history)

## Per-file changes

- `src/csk_registry/store.py:57` — `upstream_high_water` table in `_SCHEMA`
  (`key_id` PK, `version`, `log_size`, `head`, `merkle_root`, `updated_at`).
- `src/csk_registry/store.py:80-83` — schema version bump 3 → 4
  (`_SCHEMA_VERSION=4`, `_SCHEMA_VERSION_V3=3`).
- `src/csk_registry/store.py:158` — `UpstreamHighWater` dataclass
  (version/log_size not forced equal; `updated_at` is local advance time).
- `src/csk_registry/store.py:286-293` — startup migration routing
  v3→v4 and v2→v3→v4.
- `src/csk_registry/store.py:446,465` — `_require_current_schema` requires
  the new table and its `key_id` primary key.
- `src/csk_registry/store.py:535,588` — `_require_v3_schema` and
  `_migrate_v3_to_v4` (idempotent `CREATE TABLE IF NOT EXISTS` + marker bump).
- `src/csk_registry/store.py:1175,1180` — `get_upstream_high_water` /
  `_get_upstream_high_water_locked` accessor.
- `src/csk_registry/store.py:1195-1270` — `append_imports(..., upstream_high_water=None)`:
  when given, the row is written in the same `BEGIN IMMEDIATE` transaction as
  the imported records (never before commit, never partially); a concurrent
  advance past the offered boundary fails the transaction instead of lowering.
- `src/csk_registry/store.py:1524-1554` — `integrity_errors` validates every
  `upstream_high_water` row (key_id 16-hex, `0 <= log_size <= version`,
  hex heads/roots, timestamp); malformed rows fail readiness like other §5 state.
- `src/csk_registry/bundle.py:22-24` — closed diagnostics
  `import_upstream_rollback`, `import_upstream_inconsistent` (no existing
  `profiles/registry-service.md` code applies; §6 codes are restore-specific).
- `src/csk_registry/bundle.py:29,52` — `UpstreamRollbackError` /
  `UpstreamInconsistentError`, messages naming `key_id` and both boundaries.
- `src/csk_registry/bundle.py:93,149-195` — `import_bundle(..., accept_older_upstream=False)`:
  post-verification §5 comparison (below → refuse; equal+different → refuse;
  equal+identical → no-op with nothing persisted; higher → import + advance
  atomically; flag turns only below into a warning import without lowering).
- `src/csk_registry/bundle.py:230-277` — `import_bundle` audit event
  (`key_id`, `persisted`/`offered` boundaries, `result`, `diagnostic`/`warning`,
  `accepted_older`, `imported`); refusals and flag path at WARNING, ok/noop at INFO.
- `src/csk_registry/cli.py:240,258` — `_ensure_import_audit_sink` so the
  structured event reaches stderr.
- `src/csk_registry/cli.py:269,274-276` — `import-bundle` passes the flag and
  prints a human-readable warning to stderr when it is given.
- `src/csk_registry/cli.py:470` — `--accept-older-upstream` flag definition.
- `tests/test_registry.py:2030-2045` — v2 migration test updated to expect
  schema 4 (legitimate contract change from the table addition).
- `tests/test_registry.py:2880-3253` — 12 new P4 tests (see AC mapping).
- `README.md` — model bullet, CLI example, new "Upstream import high-water"
  section, schema-4 migration note.
- `SECURITY.md` — what the high-water protects against (stale-bundle replay).
- `CHANGELOG.md` — Unreleased Security entry "P4: …" (curator-spec `dced9b8`,
  implementation detail, no wire change).

## AC mapping (file:line of test + production site)

AC: "Rollback-bundle test fails closed" plus brief §Deliverable-3 (7 cases):

1. Rollback (older version) refused —
   `tests/test_registry.py:2898` (`test_import_rollback_bundle_refused`);
   production `src/csk_registry/bundle.py:151-154`.
   Diagnostic names `key_id` + both heads; high-water and log untouched.
2. Same version different body refused even with flag —
   `tests/test_registry.py:2925`
   (`test_import_inconsistent_bundle_refused_even_with_flag`);
   production `src/csk_registry/bundle.py:155-162`.
3. Identical re-import no-op —
   `tests/test_registry.py:2965` (`test_identical_reimport_is_noop`);
   production `src/csk_registry/bundle.py:163-172`.
   Returns 0, high-water byte-identical, `result="noop"` in the audit event.
4. Newer bundle imports and advances —
   `tests/test_registry.py:2993` (`test_newer_bundle_advances_high_water`);
   production `src/csk_registry/bundle.py:187-195` +
   `src/csk_registry/store.py:1253-1269` (same-transaction advance).
5. Flag imports older with warning, high-water untouched —
   `tests/test_registry.py:3013`
   (`test_accept_older_upstream_imports_without_lowering`, fork case with
   1 genuinely new record) and CLI `tests/test_registry.py:3149`
   (`test_import_bundle_cli_flag_path`, exit 1 → 0, stderr warning);
   production `src/csk_registry/bundle.py:175-186` (no high-water arg) +
   `src/csk_registry/cli.py:274-276`.
6. Failed import leaves state untouched —
   `tests/test_registry.py:3058` (chain break via swapped records) and
   `tests/test_registry.py:3086` (injected trigger failure mid-transaction
   proving the high-water rolls back with the records);
   production ordering `src/csk_registry/bundle.py:52-75` (verify before
   compare) + transactional `src/csk_registry/store.py:1218-1269`.
7. First import establishes —
   `tests/test_registry.py:2880`
   (`test_first_import_establishes_upstream_high_water`);
   production `src/csk_registry/bundle.py:187-195` with `persisted is None`.
8. Per-key isolation (supporting the "per-upstream" clause) —
   `tests/test_registry.py:3112` (`test_upstream_high_water_is_per_key`).
9. Migration + backup (supporting) —
   `tests/test_registry.py:3199` (v3→v4) and
   `tests/test_registry.py:3234` (backup preserves the table).

## Backup / verify-backup choice (as the brief requires a stated choice)

Included in `backup`, not compared by `verify-backup`. `backup` uses the
SQLite backup API, so the `upstream_high_water` table copies automatically
(proven by `test_backup_preserves_upstream_high_water`). `verify-backup`
does not compare it because the signed `registry-snapshot-v1` checkpoint
carries no upstream state and the brief forbids wire/spec changes; adding a
comparison would require a new checkpoint field.

## Validation transcripts

Shell: bash, workdir the Story worktree, `set -o pipefail`.

```
$ export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
$ /tmp/csk-venv/bin/python -m pytest -q
........................................................................ [ 40%]
........................................................................ [ 80%]
..................................                                       [100%]
178 passed, 2 warnings in 78.09s (0:01:18)
PYTEST_EXIT:0
```

(Focused new-test run:
`-k "high_water or rollback or inconsistent or reimport or newer_bundle or
accept_older or failed_import or failed_append or per_key or cli_flag or
v3_database or backup_preserves"` → 14 passed, exit 0. Baseline before the
change was 166 passed.)

```
$ /tmp/csk-venv/bin/python -m mypy
Success: no issues found in 14 source files
MYPY_EXIT:0
```

No linter is configured in this repo (no ruff/flake8 config; CI gates are
pytest + mypy strict + build + docker); mypy strict is clean. The hosted
`scripts/remote-gate.sh` gate is not run by the producer (runtime runs it at
handoff).

## Deliberately out of scope

Client-side (curator) behaviour, R4–R8 tasks, spec edits, mirror-group
comparisons (S2 is client-side), Docker/deploy changes. No protocol-suite pin
move (no new vectors consumed).

## Spec gaps

None. `profiles/registry-service.md` §6 defines only restore-checkpoint
diagnostics; nothing names the import-upstream rollback/inconsistent cases,
so the brief's working spellings `import_upstream_rollback` and
`import_upstream_inconsistent` are kept as implementation-detail diagnostics
(no spec patch per the brief's "no spec edits").
