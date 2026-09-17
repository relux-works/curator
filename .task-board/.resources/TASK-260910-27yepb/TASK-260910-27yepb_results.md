# TASK-260910-27yepb results — cursor bound to the page boundary (R1/P1 service half, second leaf)

Story `STORY-260910-3rvvxh`, spec curator-spec `dced9b8`. Builds on the
accepted `TASK-260910-14dnb7` rev 2 checkpoint (cursor carries the complete
signed chain boundary; see its results resource for the payload shape, reused
unchanged here). Worktree:
`curator-skill-registry/.temp/STORY-260910-3rvvxh/worktree` (branch
`task-board/story/STORY-260910-3rvvxh`), changes left uncommitted for handoff.

## Per-file changes

- `src/csk_registry/store.py:125-126`: new `CursorBoundaryMismatch(ValueError)`
  — a page requested at a boundary the store disagrees with.
- `src/csk_registry/store.py:577-595` (`records_page`) and `:631-646`
  (`log_page`): new keyword-only `boundary: SnapshotBoundary | None`. Cursor
  pages pass the cursor's carried boundary; `max_seq` + `boundary` together is
  a programming error (`ValueError`).
- `src/csk_registry/store.py:678-706` (`_page_boundary_locked`): resolves the
  paging cap under the store lock. A carried boundary is re-derived from the
  store at its `log_size` and compared as a full struct
  (`version`/`log_size`/`head`/`merkle_root`/`created_at`); a forged body, an
  unavailable size, or a pruned/non-contiguous prefix raises
  `CursorBoundaryMismatch` — never a silent re-evaluation at a newer boundary.
- `src/csk_registry/app.py:242-275` (`GET /v1/records`) and `:304-335`
  (`GET /v1/log`): cursor pages pass `boundary=boundary` (the cursor's carried
  boundary from `_cursor_state`) into the store call; `CursorBoundaryMismatch`
  maps to `404 invalid_cursor`. First pages are unchanged (`max_seq` at a
  freshly read `snapshot_boundary()`). No fallback to `snapshot_boundary()`
  exists on the cursor path.
- `src/csk_registry/app.py:472-484` (`_cursor_state` docstring): documents the
  two-layer gate — the pre-check (`boundary_available`, a signed-but-
  inconsistent object is `invalid_cursor` before anything is served or echoed)
  plus the structural re-verification inside the paging call.
- `tests/test_registry.py:967` (`test_cursor_carried_boundary_disagreement_…`):
  forged carried boundary (body `head`/`merkle_root` flipped, genuinely
  re-signed with the accepted service key) refused with `404 invalid_cursor` on
  both endpoints; genuine cursor still continues the chain (control).
- `tests/test_registry.py:1029` (`test_cursor_disagreement_refused_for_overlap…`):
  same refusal when the forged body is re-signed with the retained overlap key
  and the envelope with the active key, both endpoints.
- `tests/test_registry.py:1086` (`test_cursor_unavailable_boundary_…`): carried
  boundary past the committed head refused on both endpoints (404, never 200
  with another boundary).
- `tests/test_registry.py:1134` (`test_cursor_pruned_prefix_…`): previously
  valid cursors on both endpoints refused after the earliest log row is pruned
  behind the store (controls pass before pruning).
- `tests/test_registry.py:1188` (`test_store_page_calls_verify_carried…`):
  store contract — committed boundary pages identically to the equal `max_seq`
  cap; tampered body, future size, and double cap raise as specified, both page
  methods.
- `tests/test_protocol_conformance.py:362`
  (`test_shared_service_cursor_boundary_cases`): drives
  `pagination.cursor_boundary_cases` (`cursor-boundary-disagreement` → status
  404, `invalid_cursor`, `reevaluate_at_newer_boundary is False`) through the
  real endpoints on both `/v1/records` and `/v1/log`; the shared
  `append_after_first_page` flow is the control proving the chain is NOT
  re-evaluated (original cursor keeps the original boundary and
  `expected_original_cursor_ids` while `/v1/snapshot` moves on).
- `tests/test_protocol_conformance.py:450`
  (`test_shared_service_cursor_rejections`): drives every
  `pagination.cursor_rejections` name (`changed_query`, `changed_limit`,
  `wrong_endpoint`, `expired` via a genuinely re-signed past-expiry envelope,
  `unavailable_snapshot`) through the real endpoints, both directions where
  applicable, asserting `pagination.invalid_cursor_status` (404) +
  `invalid_cursor`.
- `CHANGELOG.md:36-50`: R1 Unreleased entry extended with the P1 cursor-binding
  sentence (fields verified, 404, no re-evaluation, both endpoints, spec
  `dced9b8`, `pagination.cursor_boundary_cases`).
- `README.md:24-31`: pagination bullet extended with the refusal sentence.
- `SECURITY.md:23-30`: fail-closed pagination sentence.

`ci.yml` protocol-suite pin already at
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (moved by the first leaf); no new
spec revision is consumed, so the pin is untouched and every pre-existing
conformance test stays green at it.

## Acceptance criteria

- Test proves cursor/boundary disagreement is rejected: YES. Forged
  carried-boundary bodies with valid accepted-key signatures are refused with
  `404 invalid_cursor` on both endpoints at unit level
  (`test_registry.py:967,1029`) and through the shared vectors at conformance
  level (`test_protocol_conformance.py:362`), with genuine-cursor controls
  proving the refusal is caused by the disagreement. Unavailable-size
  (`:1086`, `:450`) and pruned-prefix (`:1134`) variants are refused the same
  way, never re-evaluated.

## Validation transcripts

All from the worktree, `sh` with `set -o pipefail` semantics (no pipe masked
any reported status), interpreter `/tmp/csk-venv/bin/python` = Python 3.14.6
(venv outside the tree; CI runs 3.11 and 3.14 on three OSes).
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec `dced9b8`, verified via `git rev-parse HEAD`).

- `export CURATOR_CONFORMANCE_ROOT=... && /tmp/csk-venv/bin/python -m pytest -q`
  → `133 passed, 2 warnings in 3.94s`, exit 0. (Baseline at first-leaf
  checkpoint was 126; +7 new: 5 unit + 2 conformance. Warnings are the
  pre-existing starlette/anyio deprecations.)
- `/tmp/csk-venv/bin/python -m mypy` (strict, per `pyproject.toml`)
  → `Success: no issues found in 13 source files`, exit 0.
- `git diff --check` → clean, exit 0.
- `/tmp/csk-venv/bin/python -m build --outdir /tmp/csk-dist-p1` → success,
  exit 0 (sdist + wheel).
- No linter is configured in this repo (unchanged; static gate is mypy strict).

Negative proof (disposable copy `/tmp/csk-p1-neg`, removed afterwards; real
tree untouched, `git status` shows only the 7 intended files):

- Mutant A1 (both layers narrowed to size-only: field-equality deleted in
  `_page_boundary_locked` and `boundary_available` reduced to a size probe) →
  exit 1: exactly the 4 field-disagreement tests fail (forgeries admitted with
  200), while the 3 size/unavailability tests pass. Proves the `head`/`merkle`
  comparison class, not just gate presence.
- Mutant B (only the `_cursor_state` pre-check deleted, structural store gate
  intact) → exit 0, `133 passed`. Proves the store-call gate independently
  refuses every forgery/unavailability case.
- Mutant C (records cursor pages re-evaluated at the current boundary and
  re-signed) → exit 1: 5 failures, including the new
  `test_shared_service_cursor_boundary_cases` append control and the
  pre-existing chain/rotation tests. Proves no-reevaluation is pinned.

## Deliberately out of scope

- R2 boundary memoization/caching, the manager client, spec edits,
  tags/releases, Docker/deploy changes, key-rotation policy changes.

## Spec gaps found

None. Observation (not a gap): the `cursor_rejections` vector names were
previously undriven by any test; they are now all driven through the real
endpoints by `test_shared_service_cursor_rejections`, so "existing cases
green" is a measured claim (10 refusal probes, all 404 `invalid_cursor`).
