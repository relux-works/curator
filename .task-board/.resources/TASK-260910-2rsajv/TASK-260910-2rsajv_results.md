# TASK-260910-2rsajv results — service-idempotency-ttl-slack (R8)

Story `STORY-260910-2xe3n2`, wave 4, leaf 2 of 4. Built on the checkpointed
`TASK-260910-28kmef` tree (branch `task-board/story/STORY-260910-2xe3n2` at
`d7f424c` + uncommitted R8 change). Worktree:
`<control-root>/.temp/STORY-260910-2xe3n2/worktree`.

## Change summary (4 files, +72/−2)

- `src/csk_registry/app.py:65-70` — `IDEMPOTENCY_TTL_SECONDS = 26 * 3600`
  with a comment quoting the profile minimum verbatim ("Retention is at
  least 24 hours from the first successful commit and is not shortened by
  restart, cleanup, or credential rotation.", `profiles/registry-service.md`
  §4 at curator-spec `dced9b8`) plus the two-hour slack rationale. Consumed
  unchanged at `src/csk_registry/app.py:461` (`submit` →
  `store.append_idempotent(ttl_seconds=...)`).
- `README.md:194` — operator upgrade guidance now says the prior **26-hour**
  retention window.
- `CHANGELOG.md:36-43` — Unreleased `### Security` entry "R8: idempotency
  retention raised from 24 h to 26 h …" (behavior change, unchanged
  envelopes, curator-spec `dced9b8`).
- `tests/test_registry.py:1560-1618` — `_ManualAppClock` (injectable `time`
  replacement, same pattern as the suite's `_ManualStoreClock`) plus two
  endpoint-level tests driving `POST /v1/records` through the real app:
  - `test_idempotency_retry_within_slack_window_is_deduplicated` (:1573):
    first submit → 201, retry at +25 h → 200 with the identical body, log
    still 1 entry.
  - `test_idempotency_retry_past_retention_expires` (:1596): first submit →
    201, retry at +26 h + 1 s → 201 with `seq+1`, log 2 entries.

## How each AC line is met

1. "Constant change with a comment citing the profile minimum" —
   `src/csk_registry/app.py:65-70`: `26 * 3600`, comment quotes §4 at the
   pinned `dced9b8` and states the slack rationale.
2. "Existing idempotency tests adjusted" — no adjustment needed: the
   pre-existing tests pass `ttl_seconds` (86400 / 24·3600) directly to
   `Store.append_idempotent` and exercise store mechanics (auditor scope,
   conflict, rollback), all still valid since the 24 h profile floor in
   `src/csk_registry/store.py:946-947` is unchanged. None referenced the app
   constant. Verified: all 7 idempotency-selected tests pass.
3. "Retry between 24 h and 26 h deduplicated" —
   `tests/test_registry.py:1573` (+25 h → 200 replay, no second append).
4. "Retry at > 26 h expires" — `tests/test_registry.py:1596` (+26 h + 1 s →
   201 new append).
5. "CHANGELOG Unreleased entry R8" — `CHANGELOG.md:36-43`.

## Validation transcripts (all from the Story worktree)

Interpreter: `/tmp/csk-venv/bin/python`, Python 3.14.6 (editable install
resolving to this worktree; `csk_registry.app` = worktree `src`, TTL 93600
confirmed pre-run). Conformance root: `CURATOR_CONFORMANCE_ROOT=
/tmp/spec-47c3c8c/conformance/v1` (pin `47c3c8c`, per CI). Shell: bash,
`set -o pipefail` semantics via `PIPESTATUS`.

- `export CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 &&
  /tmp/csk-venv/bin/python -m pytest -q` → **177 passed**, exit **0**
  (164.53 s; 2 pre-existing starlette/anyio deprecation warnings only).
- `/tmp/csk-venv/bin/python -m mypy` (strict per `pyproject.toml`) →
  `Success: no issues found in 15 source files`, exit **0**.
- Narrowing mutant (negative control, reverted afterwards): TTL temporarily
  set back to `24 * 3600` → `test_idempotency_retry_within_slack_window_
  is_deduplicated` **FAILS** (replay becomes a double-append at +25 h),
  `test_idempotency_retry_past_retention_expires` still passes; constant
  restored to `26 * 3600` and diff re-verified (`git diff --stat`: 4 files,
  +72/−2). So the test pair pins retention to (25 h, 26 h + 1 s], i.e. 26 h
  at hour granularity.
- No lint config beyond mypy in this repo; no separate build step
  (`pip install -e '.[dev]'` exit 0 doubles as the compile check; pytest
  imports the full package).

## Deliberately out of scope

- R5 (landed by `TASK-260910-28kmef`), R7 (`TASK-260910-3u9t1e`) and R4
  (`TASK-260910-2c7s0u`) follow as separate leaves.
- The historical `CHANGELOG.md` line about the 24-hour legacy-drain and the
  `store.py` 24 h floor were intentionally left untouched: legacy rows were
  written under the old TTL (24 h drain remains accurate for that one-time
  migration), and the floor enforces the profile *minimum*, not the
  retention.
- `SECURITY.md` mentions no retention hours; no docstrings document the
  retention value (verified by grep for `24.hour|24h|86400|retention`).

## Spec gaps

None found. Profile §4 sentence at `dced9b8` matches the audit's
characterization of the minimum exactly.
