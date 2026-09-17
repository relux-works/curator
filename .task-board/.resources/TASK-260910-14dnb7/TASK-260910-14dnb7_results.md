# TASK-260910-14dnb7 results — service-boundary-response-fields (R1 service half)

Story `STORY-260910-3rvvxh`, spec curator-spec `dced9b8`.
Worktree: `curator-skill-registry/.temp/STORY-260910-3rvvxh/worktree` (branch
`task-board/story/STORY-260910-3rvvxh`), changes left uncommitted for handoff.

## Per-file changes

- `src/csk_registry/app.py:260-261` (`GET /v1/records`): success envelope is
  now `{"records", "next_cursor", "boundary"}` where `boundary` is
  `build_snapshot(store, signing_key, boundary=boundary)` — the full signed
  `registry-snapshot-v1` object (all fields incl. `sig`) for the committed
  boundary the page was evaluated at. First pages use
  `store.snapshot_boundary()`; cursor pages re-derive the identical signed
  object from the cursor-carried boundary (immutable `created_at` preserved;
  never a refreshed timestamp).
- `src/csk_registry/app.py:310-311` (`GET /v1/log`): same for
  `{"entries", "next_cursor", "boundary"}`.
- `tests/test_registry.py:782-853`: `test_records_pages_carry_byte_identical_boundary`
  and `test_log_pages_carry_byte_identical_boundary` — closed 3-member
  envelope on first and cursor pages, CCJ-1 byte-identical `boundary` across
  the whole chain, `verify_signed` against the service key, equality with
  `/v1/snapshot` for the same boundary, and chain pinned to the original
  boundary (incl. `created_at`) after a concurrent append.
- `tests/test_protocol_conformance.py:170-296`: `_spec_schema` /
  `_check_page_envelope` mirror of `records-response-v2` / `log-response-v2`
  loaded from `CURATOR_CONFORMANCE_ROOT/../../schemas/v1/` and cross-checked
  against the schema files' own `required` + `additionalProperties: false`;
  `test_shared_service_page_envelope_schema_cases` asserts the registered
  `schema-cases/index.json` valid instance passes and the invalid instance
  (missing `boundary`) fails the same checker; `test_shared_service_page_boundary_vectors`
  drives `registry-service.json` `pagination.boundary_emitted_on_every_page`
  and `pagination.chain_boundary_byte_identical` through the real HTTP
  endpoints, asserts `boundary_log_size`, signature verification, equality
  with `/v1/snapshot` at that boundary, the `append_after_first_page` vector
  (cursor page keeps the original boundary while a fresh query moves on), and
  a full `/v1/log` cursor chain with byte-identical boundaries.
- `.github/workflows/ci.yml:30`: protocol-suite `ref` moved
  `cbe912d` → `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (settled pin).
- `CHANGELOG.md:36-40`: Unreleased `R1:` entry naming the `boundary` member,
  the v2 envelope schemas, and spec `dced9b8`.
- `README.md:24-28,57-64`: Model + Endpoints mention of the signed per-page
  `boundary`.

Cursor semantics (`_encode_cursor` / `_cursor_state`) untouched: no
disagreement refusal added — that is the sibling task `TASK-260910-27yepb`.
Re-signing cursor pages with the active key yields byte-identical boundaries
while the key is stable; across a rotation the sig necessarily rotates with
the key (spec §5: rotation MAY replace only the outer signature).

## Acceptance criteria

- Responses carry the boundary; shared conformance vector passes: YES.
  Every `/v1/records` GET and `/v1/log` success envelope carries REQUIRED
  `boundary`, byte-identical across a cursor chain incl. after appends
  (app.py:260-261,310-311); conformance vectors
  `pagination.boundary_emitted_on_every_page` /
  `pagination.chain_boundary_byte_identical` driven through the real endpoints
  (test_protocol_conformance.py:214-296); envelopes checked against the v2
  schemas and their registered schema-cases (test_protocol_conformance.py:177-212).

## Validation transcripts

All commands run from the worktree
(`curator-skill-registry/.temp/STORY-260910-3rvvxh/worktree`), `sh` with
`set -o pipefail`, interpreter `/tmp/csk-venv/bin/python` = Python 3.14.6
(venv outside the tree per campaign rules; CI runs 3.11 and 3.14 on three
OSes). `CURATOR_CONFORMANCE_ROOT` =
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec `dced9b8`, verified via `git rev-parse HEAD`).

- `export CURATOR_CONFORMANCE_ROOT=... && /tmp/csk-venv/bin/python -m pytest -q`
  → `124 passed, 2 warnings in 2.44s`, exit 0. (The 2 warnings are
  pre-existing `starlette.testclient`/anyio deprecation warnings, also present
  at baseline.)
- Focused new tests (2 unit + 1 vector + 2 schema-case params) → `5 passed`,
  exit 0.
- `/tmp/csk-venv/bin/python -m mypy` (strict, per `pyproject.toml`)
  → `Success: no issues found in 13 source files`, exit 0.
- `/tmp/csk-venv/bin/python -m build --outdir /tmp/csk-dist-check` → success,
  exit 0; `twine check` on both artifacts → PASSED, exit 0.
- No linter is configured in this repo (no ruff/flake8/black config; the
  configured static gate is mypy strict, green as above).

Pre-existing conformance tests at the new pin: green with no changes needed.
Baseline before this change (new tests excluded) was 119 passed at
`CURATOR_CONFORMANCE_ROOT` = curator-spec `dced9b8`; after the change the full
suite is 124 passed (119 + 5 new: 2 unit + 1 vector + 2 schema-case params).
The only `registry-service.json` delta between the old pin and `dced9b8` is
the additive pagination block (`boundary_emitted_on_every_page`,
`chain_boundary_byte_identical`, `cursor_boundary_cases`), which no
pre-existing test consumes.

Negative proof: with `src/csk_registry/app.py` temporarily reverted
(`git stash push -- src/csk_registry/app.py`), the 3 service-level boundary
tests fail (`3 failed`, exit 1); with the fix restored, all pass.

## Deliberately out of scope

- Cursor/boundary disagreement refusal (`404 invalid_cursor`) and its vector
  `pagination.cursor_boundary_cases` → sibling `TASK-260910-27yepb`.
- R2 caching / memoization, the manager client, spec edits, tags/releases,
  Docker/deploy changes.

## Spec gaps found

None. One observation (not a gap): byte-identical chain boundaries assume a
stable signing key; a rotation mid-chain re-signs later pages with the new key
(allowed by registry-service §5), which a strict byte-comparing client would
read as `registry_page_boundary_mismatch`. No test pins that corner; the
existing rotation test (`test_staged_key_rotation_preserves_snapshot_body_and_live_cursors`)
still passes since cursor verification uses the overlap pin set.

## Revision 2 (rework after `changes_requested`)

Both review corrections applied; everything else from revision 1 unchanged
except where a correction touches it (noted below).

### Correction 1 — chain boundary survives staged key rotation (High)

- `src/csk_registry/app.py:406-431` (`_encode_cursor`): the cursor `snapshot`
  member is now the complete signed `registry-snapshot-v1` object (including
  `sig`) served as the page `boundary`, not the unsigned body. The cursor
  envelope itself is still signed with the active key; only the carried
  boundary is pinned to the chain.
- `src/csk_registry/app.py:434-492` (`_cursor_state`): returns
  `(offset, boundary, snapshot)`; the carried snapshot must pass
  `validate_snapshot` and `verify_signed` against the accepted key set
  (app.py:468-470), otherwise `404 invalid_cursor`. The paging boundary body
  is re-derived from the carried object via `SnapshotBoundary.from_dict`
  (app.py:471-476), so `version == log_size` is still enforced, and
  `store.boundary_available` still applies. Cursors issued before this change
  (unsigned `snapshot` body) fail `validate_snapshot` → `invalid_cursor`;
  cursors are short-lived (1h TTL), so no migration path is provided.
- `src/csk_registry/app.py:237-268` (`GET /v1/records`) and `:286-319`
  (`GET /v1/log`): first pages build the boundary with the active signer and
  issue cursors carrying exactly that object; cursor pages echo the carried
  object verbatim and re-emit it on continuations. `/v1/snapshot` still uses
  the active signer.
- `tests/test_registry.py:739-836`
  (`test_rotation_overlap_keeps_chain_boundary_on_both_endpoints`): chains on
  both endpoints started before `activate-key-rotation` and walked during the
  overlap — every page's CCJ-1 boundary bytes equal the first page's, verify
  against the OLD key, and do not verify against the new key; `/v1/snapshot`
  moves to the new signer. After `retire-key`, pre-rotation cursors AND
  overlap-issued continuations (new-signed envelope, old-signed carried
  boundary) are refused with `404 invalid_cursor` — never re-signed. Fresh
  chains after retirement serve the new signer's boundary.
- `tests/test_registry.py:664-736`: the existing staged-rotation test now also
  asserts the continued page's `boundary` equals the first page's and verifies
  against the old key (the reviewer's probe, made permanent), plus a cursor
  length bound assertion.
- Measured cursor size: 781 chars on both endpoints (probe
  `/tmp/measure_cursor.py`), 3315 chars of headroom below the 4096 bound.
- `CHANGELOG.md:36-43`: R1 entry extended with the rotation-stable chain
  sentence. No disagreement refusal added — still the sibling's scope.
- Sibling coordination note for `TASK-260910-27yepb`: cursor payload shape is
  `{"query": <hex digest of endpoint+query>, "snapshot": <full signed
  registry-snapshot-v1 incl. sig>, "offset": <int>, "expires_at": <int>}`,
  CCJ-1 bytes, envelope `payload_b64url.sig_b64url` signed with the active
  key. The disagreement check can compare the carried `snapshot` body against
  the boundary the page would otherwise be evaluated at; the carried object
  is authenticated (envelope sig + snapshot sig, both against the accepted
  pin set) before any comparison.

### Correction 2 — real Draft 2020-12 envelope validation (Medium)

- `pyproject.toml:26`: `jsonschema>=4.23` added to the `dev` extra. CI
  installs `.[dev]`, so no workflow install change was needed.
- `tests/test_protocol_conformance.py:173-198`: `_spec_schema` /
  `_check_page_envelope` replaced with `_envelope_validator`, which validates
  served envelopes against the actual `records-response-v2` /
  `log-response-v2` schemas via `Draft202012Validator` with a local
  `referencing` registry built from every `schemas/v1/*.schema.json`
  (`registry-snapshot-v1`, `registry-log-entry-v1`, `audit-record-v1`,
  `common`, …). Schema-cases (`:199-220`) now assert valid/invalid against the
  real schemas (`ValidationError`).
- `tests/test_protocol_conformance.py:250-333`: the boundary-vector test signs
  the shared vector bodies before appending (the vectors carry no `sig`, and
  `log-response-v2` requires log records to validate as `audit-record-v1`);
  the audit `case` markers the vectors assert on are preserved. Every served
  envelope on both endpoints is validated with the real schemas.
- `tests/test_protocol_conformance.py:222-248`
  (`test_shared_service_log_harness_rejects_malformed_entry_hash`): negative
  regression for the reviewer's survivor — a served log envelope validates,
  then the same envelope with `entry_hash: "invalid"` is rejected by the
  harness.

### Revision 2 validation transcripts

Same worktree, `sh` with `set -o pipefail`, interpreter
`/tmp/csk-venv/bin/python` = Python 3.14.6,
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec `dced9b8`).

- `python -m pytest -q` → `126 passed, 2 warnings in 32.24s`, exit 0
  (rev1 124 + new rotation-overlap test + new negative-regression test).
- `python -m mypy` (strict) → `Success: no issues found in 13 source files`,
  exit 0.
- `git diff --check` → clean, exit 0.
- No linter configured (unchanged from rev1; static gate is mypy strict).
- Narrowing mutants in a disposable copy (`/tmp/csk-rev2-neg`, restored
  after each): (A) served `entry_hash "invalid"` → 2 failed, exit 1 (the rev1
  survivor is now killed); (B) records cursor pages re-sign with the active
  key → both rotation tests fail, exit 1; (C) carried-snapshot signature
  check dropped → overlap test fails on the post-retire refusal, exit 1.
  3/3 killed, each by the test named for its correction.
