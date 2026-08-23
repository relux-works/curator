# TASK-260720-2dnqw2 implementation evidence

## Provenance

- Product repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Accepted base SHA: `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- Task branch: `task/TASK-260720-2dnqw2-canonical-build-metadata`
- Task worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2dnqw2/worktree`
- Dependency handoffs verified before worktree creation:
  `TASK-260720-3c0ss2`, `TASK-260720-3j8pp5`, and
  `TASK-260729-3nx97g` were all `done` with outcome evidence.
- An earlier Claude run was cancelled by operator directive and left an
  incomplete, import-broken draft in this worktree. This run audited and
  completed that inherited state without discarding it.

## Candidate conformance identity

All golden values and schema cases were read through the caller-supplied
`CURATOR_CONFORMANCE_ROOT`:

`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`

The suite is candidate-only, non-release evidence. No committed release pin or
conformance claim was changed.

- `manifest.json`:
  `sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- Canonical input: 869 bytes,
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`
- Exact receipt: 1120 bytes,
  `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`
- Legacy rc.4 input without `execution_policy`:
  `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`
- Reserved hardened input:
  `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037`
- The three derived keys are pairwise distinct. Both non-portable identities
  remain schema-invalid and cannot alias the portable key.

## Implemented scope

- Added typed `go-v1` logical input, fixed policy, cache-key, artifact, and
  receipt-schema-1 models under `src/csk/builds/metadata.py`.
- Added exact CCJ-1 value canonicalization and strict reading in
  `src/csk/protocol_json.py`; audit-registry signing behavior continues to
  remove only its own top-level signature envelope before canonicalization.
- Added dedicated typed install-marker schemas 1 and 2 in
  `src/csk/install_marker.py`, including deterministic sorting, closed
  optional members, build-source iff-builds validation, and schema-level
  legacy-currentness compatibility.
- Added focused conformance-driven tests for exact bytes and hashes, every
  supplied receipt-v1 and marker-v1/v2 schema case, duplicate/unsafe/
  noncanonical JSON, unknown fields, every expected-input mismatch, malformed
  identities, wrong derived paths, unsupported versions, and the two
  execution-policy non-alias negatives.
- Receipt v2, marker v3, conformance claims, filesystem trust, cache layout,
  compiler execution, status, and installer mutation remain outside this
  change.

## Exact gate ledger

All gates ran directly as standalone processes without `tee` or masking pipes.

1. Focused accepted-root pytest:

   `CURATOR_CONFORMANCE_ROOT=... python -m pytest tests/test_build_metadata.py tests/test_protocol_json.py tests/test_protocol_conformance.py -q`

   Exit `0`: `170 passed in 3.00s`.

2. Full accepted-root pytest:

   `CURATOR_CONFORMANCE_ROOT=... python -m pytest -q`

   Exit `0`: `835 passed, 6 skipped in 85.55s`.
   The skips are existing platform-conditional suite cases; no selected test
   failed.

3. Strict type check:

   `python -m mypy`

   Exit `0`: `Success: no issues found in 61 source files`.

4. Distribution build:

   `python -m build`

   Exit `0`: sdist and universal wheel built successfully. The archive output
   includes both `csk/install_marker.py` and `csk/builds/metadata.py`.

5. Distribution metadata:

   `python -m twine check dist/*`

   Exit `0`: wheel and sdist both `PASSED`.

6. Repository style/whitespace gate:

   `git diff --check`

   Exit `0`, no output. The project defines no separate Ruff, Black, or
   Pyflakes command.

The initial inherited-state diagnostic exited `2` during test collection
because the cancelled draft instantiated `GoBuildPolicy` before its validator
was defined. That defect was corrected before every gate above.
