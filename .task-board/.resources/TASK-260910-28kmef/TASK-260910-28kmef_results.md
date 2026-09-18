# TASK-260910-28kmef results — R5: pathological JSON returns `400 invalid_json`, never `500`

Story `STORY-260910-2xe3n2`, first of four sequential leaves. Worktree
`<control-root>/.temp/STORY-260910-2xe3n2/worktree`, branch
`task-board/story/STORY-260910-2xe3n2`, base `c7ef32c`.

## Finding

`docs/security-audit-2026-09.md` R5 (Low): `protocol.load_json` and the
canonicalization / `_validate_ccj` path in `signing.py` recurse, so a deeply
nested document raises `RecursionError`, which the generic handler reported as
`500 internal_error` instead of `400`.

Reproduced before the fix (standalone script, TestClient with
`raise_server_exceptions=False`):

- `load_json("["*5000 + "]"*5000)` → `RecursionError`
- `signing.canonical_document_bytes(<5000-deep list>)` → `RecursionError`
- `POST /v1/records` with that body → `500 internal_error`
- `GET /v1/log` with a signed 1200-deep cursor payload (3287 chars, within the
  4096 cursor limit) → `500 internal_error`

After the fix the same probes give `JSONDepthError`, `CanonicalDepthError`,
`400 invalid_json`, and `404 invalid_cursor` respectively.

## Design

Explicit nesting bound `MAX_JSON_DEPTH = 100` (`src/csk_registry/signing.py:17`),
enforced iteratively before parsing plus `RecursionError` conversion at every
canonicalization entry point as defense in depth. 100 is far above legitimate
depth (records/snapshots/cursors nest ≤ ~6) and far below the interpreter limit
(`RecursionError` observed at depth ~1000 with this code on CPython 3.14).

- Depth violations are distinct subclasses so the request layer can map them
  without changing any other mapping: `JSONDepthError(ProtocolError)`
  (`src/csk_registry/protocol.py:35`) and
  `CanonicalDepthError(CanonicalError)` (`src/csk_registry/signing.py:30`).
  Both messages name the bound (`JSON nesting exceeds maximum depth of 100`).
- `POST /v1/records` maps depth errors to `400 invalid_json`;
  every other `ProtocolError` still maps to `400 invalid_record`, and cursors
  still map to `404 invalid_cursor` (depth included, via the existing
  `ValueError`/`ProtocolError` catch).

## Per-file changes

- `src/csk_registry/signing.py` — `MAX_JSON_DEPTH`, `CanonicalDepthError`;
  `_validate_ccj` takes `depth` and rejects nesting over the bound
  (`signing.py:74-99`); `_canonical_document` converts `RecursionError` to
  `CanonicalDepthError` (`signing.py:54-64`); `verify_signed` also treats
  `RecursionError` as verify-false (`signing.py:186`); bound stated in the
  `canonical_bytes` / `canonical_document_bytes` docstrings.
- `src/csk_registry/protocol.py` — `JSONDepthError`; iterative pre-parse
  bracket scan `_check_json_depth` for `bytes` and `str`
  (`protocol.py:41-92`); `load_json` runs the scan first and converts
  `CanonicalDepthError`/`RecursionError` to `JSONDepthError`
  (`protocol.py:124-147`); `_canonicalize_validated` maps depth failures from
  `validate_record` / `validate_snapshot` to `JSONDepthError`
  (`protocol.py:150-164,214,262`); bound stated in the `load_json` docstring.
- `src/csk_registry/app.py` — `submit` maps `JSONDepthError` /
  `CanonicalDepthError` / `RecursionError` to `400 invalid_json`
  (`app.py:414-430`); the idempotency-digest `canonical_bytes` call maps depth
  failures the same way (`app.py:439-449`); `_cursor_state` additionally
  catches `RecursionError` into the existing `404 invalid_cursor`
  (`app.py:601`). No other status/code mapping changed.
- `tests/test_registry.py` — 6 new tests + `_nested_list` helper
  (`test_registry.py:293-386`; see AC mapping below).
- `README.md` — bound stated in "Production transport and limits"
  (`README.md:162-167`).
- `CHANGELOG.md` — Unreleased Security entry "R5: …" citing
  curator-spec `dced9b8` (`CHANGELOG.md:36-43`).

## AC mapping

AC: "Deep-nesting request test asserts 400."

- `test_submit_rejects_deeply_nested_json_with_invalid_json`
  (`tests/test_registry.py:348`) posts depth-101 (just over the explicit
  bound), depth-5000 (previously `RecursionError` → 500), and depth-5000 with
  `Idempotency-Key` to `POST /v1/records`; each asserts `400` +
  `invalid_json`. It also locks the preserved mapping by asserting malformed
  (non-deep) JSON still returns `400 invalid_record`.
- `POST /v1/records` (`src/csk_registry/app.py:381`) is the only body-parsing
  endpoint: `_read_request_body` has a single caller (`app.py:397`), and the
  remaining routes are all `GET` with no body reads. No other endpoint needed
  the new mapping.
- Unit: `test_load_json_rejects_over_deep_nesting` (`:300`) — bound 100
  passes, 101 rejected for lists/dicts, `str`/`bytes`, message names the
  bound, depth error stays a `ProtocolError`, brackets in strings don't
  count; `test_canonicalization_rejects_over_deep_nesting` (`:328`) —
  `canonical_document_bytes` / `canonical_bytes` reject depth 101,
  `verify_signed` returns `False` on deep input instead of raising.
- Mutant-killers for the defense-in-depth branches:
  `test_load_json_maps_recursion_error_to_depth_error` (`:319`) and
  `test_canonicalization_maps_recursion_error_to_depth_error` (`:339`)
  force `RecursionError` via monkeypatch and assert the depth mapping.
- Cursor hardening (client JSON outside the body):
  `test_deeply_nested_cursor_is_invalid_cursor_not_500` (`:376`) serves a
  signed 1200-deep cursor and asserts `404 invalid_cursor` (was 500 before).
- Store's own JSON never trips the bound: full existing suite green
  (172 passed, below), no legitimate depth near 100.

## Validation transcripts

Venv `/tmp/csk-venv-2xe3n2` (outside the tree), CPython 3.14.6.
Only 3.14/3.12 interpreters exist on this machine; CI's 3.11 leg runs at the
hosted handoff gate. `set -o pipefail` everywhere; exit codes are real.

```text
$ export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
$ /tmp/csk-venv-2xe3n2/bin/python -m pytest -q
172 passed, 2 warnings in 56.81s
PYTEST_EXIT=0
```

(2 warnings are pre-existing third-party deprecations from
`fastapi.testclient` / `starlette.testclient`, unrelated to this change.)

```text
$ /tmp/csk-venv-2xe3n2/bin/python -m mypy
Success: no issues found in 14 source files
MYPY_EXIT=0
```

(mypy strict per `pyproject.toml`.)

```text
$ /tmp/csk-venv-2xe3n2/bin/python -m build && /tmp/csk-venv-2xe3n2/bin/python -m twine check dist/*
Successfully built curator_skill_registry-0.1.0.tar.gz and ...-py3-none-any.whl
BUILD_EXIT=0
Checking dist/...whl: PASSED / Checking dist/....tar.gz: PASSED
```

(`dist/`/`build/` removed afterwards; tree holds exactly the 6 files above.
No linter is configured in this repo — CI gates are pytest, mypy, build,
docker — so mypy strict is the static gate; it is green.)

## Deliberately out of scope

R4/R7/R8 (sibling leaves), rate limiting, request size limits, spec edits.
No CI protocol-suite pin move (this task consumes no new vectors/schemas).
`store.py` / `bundle.py` / `cli.py` untouched: their canonical calls operate
on already-validated data, and their existing `ValueError` handling already
covers the new depth subclasses (`CanonicalDepthError` is a `ValueError`).

## Notes (not spec gaps)

- The brief states the CI pin is already `dced9b8`; the worktree's
  `.github/workflows/ci.yml` actually pins `47c3c8c`. Left untouched per the
  do-not-move instruction; flagged, not fixed, as it is outside R5 scope.
- `invalid_json` is not enumerated in `protocol/registry.md` §9.1, but the
  error envelope (`error-response-v1.schema.json`) accepts any stable
  snake-case code and §9.1 mandates 400 for malformed JSON, so no spec change
  is required. No wire-schema change.

---

# Revision 2 — R1 rework: canonicalization depth errors are `ProtocolError`

Rework brief `TASK-260910-28kmef_rework-rev2.md` after
`TASK-260910-28kmef_review-verdict-rev1.md` (round 1: everything passed
except R1). All revision-1 sections above still describe revision 1;
this section describes the revision-2 delta. Line numbers below are the
revision-2 tree.

## R1 fix (the required correction)

`CanonicalDepthError` is now catchable as `ProtocolError` from BOTH
canonicalization entry points, keeping `CanonicalError` / `ValueError`
compatibility and every other mapping. Design choice (as the brief
allowed): multiple inheritance plus a shared error-definition module,
because `signing.py` importing `ProtocolError` from `protocol.py`
would be a protocol↔signing import cycle (`protocol.py` already
imports `MAX_JSON_DEPTH` / `canonical_document_bytes` from
`signing.py`).

- `src/csk_registry/errors.py` (new) — import-cycle-free home of
  `ProtocolError(ValueError)` (`errors.py:12`), with a docstring
  stating why the base lives there.
- `src/csk_registry/signing.py` — `from .errors import ProtocolError`
  (`signing.py:14`); `class CanonicalDepthError(CanonicalError,
  ProtocolError)` (`signing.py:32`) with a docstring stating both
  contracts; MRO is `CanonicalDepthError → CanonicalError →
  ProtocolError → ValueError`. `canonical_bytes` / `canonical_document_bytes`
  docstrings now state the `ProtocolError` catchability
  (`signing.py:45,55`). No `except`-clause or behavior change:
  `_canonical_document` still converts `RecursionError` (`signing.py:64-67`),
  `_validate_ccj` still raises the same bound error (`signing.py:87,95`).
- `src/csk_registry/protocol.py` — `ProtocolError` definition moved to
  `errors.py`, re-exported explicitly (`from .errors import ProtocolError
  as ProtocolError`, `protocol.py:10`; the `as` form is required by mypy
  strict's explicit-re-export rule). `JSONDepthError(ProtocolError)`
  stays in `protocol.py` (`protocol.py:32`). All existing importers
  (`app.py`, `checkpoint.py`, tests) unchanged and resolving to the same
  class object (verified `protocol.ProtocolError is errors.ProtocolError`).
- `src/csk_registry/app.py` — byte-identical to revision 1. The
  `except JSONDepthError / CanonicalDepthError / ProtocolError` order in
  `submit` still maps depth failures to `400 invalid_json` before the
  generic `ProtocolError → 400 invalid_record` clause; `verify_signed`
  and the store/bundle/CLI `ValueError` handlers still catch depth
  errors via `ValueError`.
- `CHANGELOG.md` — one-token doc correction: the R5 entry cited
  curator-spec `dced9b8`; it now cites the actual CI pin `47c3c8c`
  (`CHANGELOG.md:43`), per the corrected campaign rules. No other
  revision-1 file changed.

## R1 tests (direct `ProtocolError` assertions, both entry points)

- `test_canonicalization_rejects_over_deep_nesting`
  (`tests/test_registry.py:328`) — extended: asserts
  `issubclass(CanonicalDepthError, ProtocolError)` and
  `issubclass(CanonicalDepthError, CanonicalError)`, and catches
  `ProtocolError` for actual over-depth input to BOTH
  `canonical_document_bytes` and `canonical_bytes`.
- `test_canonicalization_maps_recursion_error_to_protocol_error`
  (`tests/test_registry.py:349`, replaces the single-entry-point
  revision-1 recursion test) — parametrized over both entry points ×
  both fault-injection points (4 cases): injected `RecursionError` from
  the CCJ validation path (`_validate_ccj`) and from JSON serialization
  (`json.dumps`) is caught as both `CanonicalDepthError` and
  `ProtocolError` for each entry point.

## Validation transcripts (revision 2)

Venv `/tmp/csk-venv-2xe3n2` (outside the tree, editable install of the
worktree), CPython 3.14.6, zsh, `set -o pipefail` everywhere; exit
codes are real. Conformance root is the detached worktree at the CI
pin, per the corrected rules (pin itself untouched:
`.github/workflows/ci.yml:30` still `47c3c8c...`).

```text
$ export CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1
$ /tmp/csk-venv-2xe3n2/bin/python -m pytest -q
175 passed, 2 warnings in 117.44s
PYTEST_EXIT=0
```

(172 revision-1 tests plus 3 net-new cases: the 1 recursion test became
4 parametrized cases. The 2 warnings are the same pre-existing
third-party `fastapi`/`starlette` testclient deprecations.)

```text
$ /tmp/csk-venv-2xe3n2/bin/python -m mypy
Success: no issues found in 15 source files
MYPY_EXIT=0
```

(mypy strict per `pyproject.toml`; 15 files now, including the new
`errors.py`. An intermediate run flagged the implicit re-export, fixed
with the explicit `as` form; the transcript above is the final tree.)

```text
$ /tmp/csk-venv-2xe3n2/bin/python -m build && /tmp/csk-venv-2xe3n2/bin/python -m twine check dist/*
Successfully built curator_skill_registry-0.1.0.tar.gz and ...-py3-none-any.whl
BUILD_EXIT=0
Checking dist/...whl: PASSED / Checking dist/....tar.gz: PASSED
```

(`dist/`/`build/` removed afterwards; `git diff --check` exit 0. Tree
holds exactly the 6 revision-1 files plus new `src/csk_registry/errors.py`.)

Narrowing-mutant proof for the R1 contract (disposable copy under
`/tmp`, candidate untouched): reverting `signing.py:32` to single
inheritance `CanonicalDepthError(CanonicalError)` makes exactly the 5
R1 assertions fail (1 over-depth + 4 parametrized recursion cases),
pytest exit 1 — the committed tests enforce the `ProtocolError`
contract rather than merely the subclass.

## Out of scope (unchanged from revision 1)

R4/R7/R8, rate limiting, request size limits, spec edits, CI pin move.
