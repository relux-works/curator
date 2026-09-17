# TASK-260910-1ny7yl results — `serve --checkpoint` startup comparison (R3/P2)

Story `STORY-260910-35tbgb`, curator-spec `47c3c8c`
(`profiles/registry-service.md` §6). Worktree
`<control-root>/.temp/STORY-260910-35tbgb/worktree`, branch
`task-board/story/STORY-260910-35tbgb`, uncommitted. No spec edits.

## Per-file changes

- `src/csk_registry/checkpoint.py` (new): closed §6 diagnostics
  (`restore_below_checkpoint`, `restore_inconsistent_with_checkpoint`,
  `checkpoint_signature_invalid`, posture `checkpoint_not_configured`,
  `REFUSAL_DIAGNOSTICS`), `CheckpointView` (no forced
  `version == log_size`, unlike `SnapshotBoundary`), and the pure
  comparator `compare_checkpoint` (`checkpoint.py:71`): below → below
  diagnostic; equal version with different `head`/`merkle_root`/`log_size`
  → inconsistent; above → serve only if the live prefix at the checkpoint
  `log_size` reproduces head and Merkle root.
- `src/csk_registry/__init__.py:11`: `CHECKPOINT_ENV =
  "CURATOR_SKILL_REGISTRY_CHECKPOINT"` (env form of the flag, same pattern
  as the other serve options).
- `src/csk_registry/store.py`: `HealthVerdict.code` (`store.py:166`,
  default `"not_ready"`); `Store._health_code` threaded through
  `health_verdict` and `refresh_health_verdict` (`store.py:1373,1481,1525,1556`);
  `refuse_startup_checkpoint` (`store.py:1673`, rejects unknown diagnostics,
  latches via the common `_latch_integrity_failure_locked`);
  `apply_startup_checkpoint` (`store.py:1689`, fetches live + R2 prefix
  boundary under lock, compares, latches on refusal, returns
  diagnostic-or-None).
- `src/csk_registry/app.py`: `_enforce_startup_checkpoint`
  (`app.py:653`) called from `app_from_env` (`app.py:744`) after `Store()`
  (§5) and before `create_app` return, hence before bind/ready; signature
  first against `public_keys(home)` (staged-rotation set), then
  comparison; `startup_checkpoint` audit events (`app.py:719`);
  `/health` serves `verdict.code` (`app.py:213`) — `not_ready` for
  integrity failures (unchanged), the refusal diagnostic for §6 refusals.
- `src/csk_registry/cli.py`: `serve --checkpoint PATH` (`cli.py:483`;
  help documents the env fallback); `_cmd_serve` exports the flag to
  `CHECKPOINT_ENV` (`cli.py:352`), so `app_from_env` is the single
  enforcement path for flag and env forms.
- `tests/test_protocol_conformance.py`: `test_shared_service_startup_checkpoint_cases`
  (`:949`) + `_run_startup_checkpoint_case` (`:769`) drive all 7
  `checkpoint_cases` through `app_from_env` + HTTP.
- `tests/test_registry.py`: `_startup_home` (`:671`); CLI flag/env test
  (`:707`); AC restore test (`:753`); malformed-checkpoint startup failure
  (`:837`); comparator closed-diagnostic test (`:856`); latch-durability
  test (`:926`).
- `.github/workflows/ci.yml`: protocol-suite `ref` →
  `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`.
- `README.md` ("Startup checkpoint gate" section + CI vector list),
  `SECURITY.md` (restore paragraph), `compose.yaml` (commented checkpoint
  mount + env), `CHANGELOG.md` (`:36`, R3/P2 entry citing
  curator-spec `47c3c8c`).

## Acceptance criteria

- AC "Test restores an older database and proves serve refuses":
  `tests/test_registry.py:753` builds a 10-record log with a checkpoint at
  10, restores a 7-record `backup_to` over `registry.db`, and starts via
  the real `app_from_env` path: `/health` → 503
  `restore_below_checkpoint`, writes → 503 `storage_unavailable`, reads
  still 200 (§5-latch parity), audit event carries the compared
  checkpoint/live boundaries + diagnostic, and the restored file is
  byte-identical afterwards (no truncate/repair).
- Brief deliverable 1 (comparison rule): `checkpoint.py:71`,
  `store.py:1689`, `app.py:653`; signature-first ordering at
  `app.py:698`; diagnostics in the startup log (`app.py:719`) and the
  `/health` error envelope (`app.py:213`).
- Brief deliverable 2 (posture): `app.py:667` records
  `checkpoint_not_configured`; `health-response-v1` unchanged (success
  envelope untouched; refusal is a 503 error-envelope code).
- Brief deliverable 3 (`verify-backup` stays; docs say which is
  normative): `cli.py:293` untouched; README "Startup checkpoint gate",
  SECURITY.md restore paragraph.
- Brief deliverable 4 (conformance + unit tests; recovery/R2 green):
  `test_protocol_conformance.py:949` (7/7 cases, names pinned);
  `test_registry.py:707,753,837,856,926`; full suite 161 passed
  including the pre-existing `recovery_cases` and R2 health tests.
- Brief deliverable 5 (CI pin): `ci.yml` `ref` =
  `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`; suite green against the
  `47c3c8c` checkout.
- Brief deliverable 6 (CHANGELOG + compose): `CHANGELOG.md:36`,
  `compose.yaml` commented example.

## Validation transcripts (exit codes real, no pipes)

Shell: bash. Interpreter: `/tmp/csk-venv/bin/python`, Python 3.14.6
(venv outside the tree per producer rules; CI covers 3.11/3.14 × 3 OS).
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec checkout at `47c3c8c`).

- `python -m pytest -q` → `161 passed, 2 warnings in 30.09s`, exit 0.
  (Baseline before changes: `155 passed`, exit 0. New: 1 conformance +
  5 unit tests. The 2 warnings are pre-existing third-party
  deprecations: fastapi/starlette testclient + anyio alias.)
- `python -m mypy` (strict per `pyproject.toml`) →
  `Success: no issues found in 14 source files`, exit 0.
- `python -m build` + `python -m twine check dist/*` → both
  distributions PASSED, exit 0 (build artifacts removed afterwards;
  worktree contains only the 11 modified + 1 new source/test/doc file).
- No linter is configured in this repo (no ruff/flake8 config; CI runs
  pytest, mypy, build, docker); mypy strict is the static gate and is
  green. The hosted gate (`scripts/remote-gate.sh`) is not run by the
  producer; it runs once at handoff.

## Deliberately out of scope

P4 import high-water, R4–R8, spec edits, key-rotation changes,
Docker/deploy changes beyond the compose example comment.

## Decisions (logbook)

- Reads keep serving during a refusal (only `/health` 503 + writes 503
  refuse), matching the §5 latch exactly: the brief defines refusal as
  "`/health` 503 through the common integrity path, writes disabled".
- A missing/unreadable/malformed checkpoint file fails startup loudly
  (`RuntimeError`, process never binds) instead of mapping to a refusal
  diagnostic, since the spec closes the diagnostic set and a malformed
  file is operator config error, not a restore condition. Only a
  well-formed snapshot with a bad signature is
  `checkpoint_signature_invalid`.
- The equal-version and above-checkpoint comparisons follow the spec
  letter (`head`/`merkle_root`/`log_size`; `created_at` is not compared),
  diverging from `Store.checkpoint_matches`, which also requires
  `created_at` equality.

## Spec gaps

None found. The `prefix_reproduced` / `same_boundary_body` vector fields
are branch-scoped hints (consulted only for live-above / live-equal per
the spec validator oracle `expected_checkpoint_verdict`); the test
measures each only on its branch.
## Rework after rev1 hosted gate failure (run 35261471138)

Hosted gate failed 1 of 6 lanes (`Tests / Python 3.14 on ubuntu-latest`) in
the R2 test `test_health_stale_verifier_fails_closed_and_refresh_recovers`
at the "age == bound is still fresh" step (`/health` 503 vs 200). Checkpoint
tests passed on every lane. Harness-only fix, no production changes; the
orchestrator's thread-race hypothesis was investigated and disproven (below),
and the real mechanism was reproduced and fixed.

- Corrected diagnosis (evidence, not the thread): the test anchored its
  manual clock to the raw `time.monotonic()` value, whose magnitude is host
  uptime. When `base + 20.0` crosses a binary binade boundary (uptime within
  20 s below a power of two) and rounds up, `(base + 20.0) - base` exceeds
  `20.0` by ~1 ulp, so the `>` staleness comparison reads stale → 503 at
  the exact-bound step. Lane-dependent, rare, deterministic given the host
  clock — matching the single-lane failure. The thread hypothesis cannot
  hold: the test uses a bare `TestClient` (no `with`), which never runs the
  serving lifespan, so `_verifier_thread` is `None` even after requests
  (probed through the real `_health_client` path; the suite itself asserts
  `not health_verifier_running()` for bare clients at `:2324`). With no
  thread, no writer can move `_health_last_refresh_monotonic` between the
  base read and the assert, leaving float rounding as the only mechanism.
- Directed repro (throwaway, `/tmp/repro_bound.py`, real `/health` path, no
  thread): `last_refresh` re-anchored to the rounding-up base `2028.001`,
  exact-bound GET → `age=20.000000000000227` → HTTP 503. Reproduces the CI
  symptom exactly under the old anchoring pattern.
- Fix (`tests/test_registry.py` only): stop the verifier before installing
  the manual clock (`:2047-2048`, ordered guard; asserts the no-thread
  premise first so a future harness change fails loudly instead of
  flaking), install the clock at the fixed binade-safe
  `_MANUAL_CLOCK_EPOCH = 10000.0` (`:2003`), re-anchor `last_refresh`
  through the production `refresh_health_verdict()` entry point, and read
  `base` after (`:2049-2055`, `assert base == _MANUAL_CLOCK_EPOCH`). Both
  exact-bound additions stay inside binade `[8192, 16384)`, where they are
  exact and the age subtraction is exact by Sterbenz. Exact-boundary
  assertions kept; no sleeps, no skips, no widened bounds. Sibling
  `_ManualStoreClock` test (`:2090`, stale side only) is robust to
  ±ulp error and untouched.
- Premise lock: `test_manual_clock_epoch_keeps_exact_bound_additions_exact`
  (`:2016`) machine-checks that both exact-bound additions read exactly
  `20.0`, so a future epoch change to a hostile value fails fast.
- Fresh transcripts (same shell/interpreter/venv as above, exit codes real,
  no pipes): `python -m pytest -q` → `162 passed, 2 warnings in 26.65s`,
  exit 0 (161 + 1 guard test; warnings are the same pre-existing
  third-party deprecations); `python -m mypy` → `Success: no issues found
  in 14 source files`, exit 0; `python -m build` + `twine check dist/*` →
  both PASSED, exit 0. Worktree still contains only the 11 modified + 1
  new file (build artifacts git-ignored, none leaking into the diff).

## Revision 3 — rework after changes_requested (review rev2)

Three P2 corrections from `TASK-260910-1ny7yl_review-verdict-rev2.md`, all
implemented as written. Everything else is unchanged: the production
comparison/latch code is untouched; the delta is one additive logging sink,
four new tests plus one comparator case, and two doc sentences.

### Correction 1 — startup posture/outcome events reach a real sink

- Root cause (confirmed): `serve` built the app — running the §6
  enforcement — before Uvicorn configured logging; `csk_registry.audit` had
  no handler and inherited WARNING, so the INFO posture/success events were
  dropped in production (reviewer's probe: 0/2).
- Fix: `src/csk_registry/app.py:653` `ensure_startup_audit_sink()` sets the
  audit logger to INFO and attaches a structured stderr sink (`%(message)s`,
  the JSON body preserved) exactly once (idempotent); it is called from
  `src/csk_registry/cli.py:357` in `_cmd_serve`, after the env export and
  before `app_from_env()`, hence before the enforcement runs. Refusal
  events keep their WARNING level and shape; propagation is unchanged, so
  the `caplog` tests are unaffected.
- Tests (`tests/test_registry.py`): `:759` resolves the real console entry
  (`curator-skill-registry` script, `sys.executable -c` fallback); `:779`
  runs it in a fresh process on an ephemeral port with a bounded terminate
  (an early exit fails loudly instead of passing silently).
  `test_serve_subprocess_records_no_checkpoint_posture_on_stderr` (`:828`)
  asserts the `checkpoint_not_configured` event on stderr with the listener
  up; `test_serve_subprocess_records_successful_comparison_on_stderr`
  (`:841`) asserts the `result: ok` event with the compared
  checkpoint/live boundaries.
- Docs: the README startup-gate paragraph notes the stderr sink and
  levels; the CHANGELOG entry notes "(stderr, structured)".

### Correction 2 — discriminating Merkle-root negative

- `tests/test_registry.py:1123` extends the comparator closed-diagnostics
  test with a root-only-diverged prefix (genuine head, different root) →
  `restore_inconsistent_with_checkpoint`.
- `test_startup_checkpoint_above_with_root_only_mismatch_refuses` (`:1136`):
  a live 10-record log with a signed checkpoint at version/log_size 8
  carrying the genuine prefix head but a tampered `merkle_root` (re-signed
  with the registry key, signature verified in-test); startup asserts
  `/health` 503 `restore_inconsistent_with_checkpoint`, writes 503
  `storage_unavailable`, and intact history (reopened head 10, ready).
- Narrowing mutant (drop `and prefix.merkle_root == checkpoint.merkle_root`
  at `checkpoint.py:100`, disposable copy `/tmp/csk-mutant-root`, same venv):
  `2 failed, 11 passed`, exit 1 — killed by both new tests (comparator
  `assert None == 'restore_inconsistent_with_checkpoint'` and startup
  `assert 200 == 503`). Previously this mutant survived all 162 tests.

### Correction 3 — the restore scenario traverses the CLI

- `test_serve_cli_entry_refuses_restored_older_database` (`:954`): restores
  the 7-record backup over the 10-record live database, invokes the real
  `main(["--home", …, "serve", "--checkpoint", …])` with the factory and
  enforcement real, intercepting only `uvicorn.run` to inspect the
  constructed app; asserts the flag→env wiring, that the app is a real
  `FastAPI`, `/health` 503 `restore_below_checkpoint`, write 503
  `storage_unavailable`, reads still 200, byte-identical history, and a
  reopened head of 7 reporting ready. The existing env-fallback and
  flag-precedence test is untouched.

### Revision 3 validation transcripts (exit codes real, no pipes)

Shell bash, `set -o pipefail`, from the worktree.
Interpreter `/tmp/csk-venv/bin/python`, Python 3.14.6 (venv outside the
tree; CI covers 3.11/3.14 × 3 OS).
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec checkout at `47c3c8c`).

- `python -m pytest -q` → `166 passed, 2 warnings in 89.47s`, exit 0
  (162 + 4 new tests; the warnings are the same pre-existing third-party
  deprecations: fastapi/starlette testclient + anyio alias).
- `python -m mypy` (strict per `pyproject.toml`) →
  `Success: no issues found in 14 source files`, exit 0.
- `git diff --check` → exit 0. Worktree still contains only the 11
  modified + 1 new file.
- Root-drop mutant (above) → exit 1 with exactly the two expected
  failures.
- Reviewer-pattern subprocess probe (`/tmp/csk-rev3-evidence/probe.py`,
  installed console entry, empty store, bounded 8 s runs terminated after
  startup): both the no-checkpoint and the valid-checkpoint runs emit the
  `startup_checkpoint` JSON on stderr before the Uvicorn lines — 2/2
  present (was 0/2 at rev2):

```text
absent exit -15
{"configured":false,"event":"startup_checkpoint","posture":"checkpoint_not_configured"}
INFO:     Started server process [47203]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:60267 (Press CTRL+C to quit)
INFO:     Shutting down
INFO:     Waiting for application shutdown.
INFO:     Application shutdown complete.
INFO:     Finished server process [47203]

startup_checkpoint present: True
============================================================
valid exit -15
{"checkpoint":{"head":"0000000000000000000000000000000000000000000000000000000000000000","log_size":0,"version":0},"configured":true,"event":"startup_checkpoint","live":{"head":"0000000000000000000000000000000000000000000000000000000000000000","log_size":0,"version":0},"result":"ok"}
INFO:     Started server process [47449]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:60269 (Press CTRL+C to quit)
INFO:     Shutting down
INFO:     Waiting for application shutdown.
INFO:     Application shutdown complete.
INFO:     Finished server process [47449]

startup_checkpoint present: True
============================================================
```

The hosted gate (`scripts/remote-gate.sh`) is not run by the producer; it
runs once at handoff.
