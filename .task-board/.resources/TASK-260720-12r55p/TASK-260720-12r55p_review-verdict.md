# TASK-260720-12r55p reviewer verdict — CHANGES REQUESTED

Reviewed CocoaSkills PR 19 at exact signed head `b754bd7aeba87baca0c63435ddc6a14d80c29400` against base `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09` and the full updated task acceptance criteria.

## Verdict

CHANGES REQUESTED. Route to `to-dev`; this is ordinary implementation and test-harness rework, not a stop-the-line boundary.

## Blocking findings

1. **P1 — 75 of 77 build-driver rejection vectors do not assert CocoaSkills behavior.** In `tests/protocol_conformance_adapters.py:138`, `assert_build_rejection_case` validates only that vector metadata says reject and returns at lines 147–149 whenever `input.build_input` is absent. The reviewed suite has 77 cases: only 2 carry `input.build_input`; 73 carry only a declarative condition and 2 use other shapes. An in-memory mutation of `unknown-driver.expected.error` to `definitely-not-the-product-error` still passed. This does not satisfy independent named negative coverage. Add an exhaustive vector-name or boundary adapter that drives the corresponding CocoaSkills seam, consumes vector-provided values, and asserts the vector error, no reuse, no execution, and fail-closed boundary. Unknown or unbound cases must fail.

2. **P1 — fixed environment and all five Go argv forms are not covered byte-for-byte.** The adapter at `tests/protocol_conformance_adapters.py:491` checks only 3 of 28 environment entries and only partial suffixes for 3 argv records. It does not independently assert telemetry-off, version, every cwd/source-aware flag, or the full environment. A mutation to `GOENV=poisoned` plus `go telemetry on` still passed. Capture the actual parent probe and worker plan from CocoaSkills and compare the complete environment and all five direct argv records to the caller-supplied vector.

3. **P1 — expanded manager lifecycle assertions are largely self-descriptions, not implementation conformance.** For example, the transaction branch at `tests/protocol_conformance_adapters.py:796` ignores `manager_home_lock_held_through_rollback` and `require_current_digest_equals_desired_before_restore`; setting both false still passed. Publication, cross-project, GC, private-build, recovery, repair, status, and rollback branches similarly mostly compare fields within the vector or to literals without driving CocoaSkills behavior. Bind each of the 32 cases to a product seam or reusable existing test helper and assert every normative field; keep the mapping exhaustive and fail on unknown fields/cases.

4. **P1 — the caller-supplied root is not authenticated as the exact manifest content set.** `tests/test_protocol_conformance.py:95` hashes only `manifest.json`. Lines 288–296 verify manifest hashes only for the 11 build-driver goldens; the consumed schema cases and four required vector JSON files are never checked against their manifest entries. Therefore a root can retain the required manifest digest while serving altered vector or schema bytes, and the self-asserting adapters can still pass. Verify the full in-scope artifact inventory and every consumed artifact digest from the pinned manifest before loading cases.

5. **P1 — several minimum build-source and toolchain cases also stop at vector metadata.** The missing-LF and multiple-LF toolchain vectors publish `input.stdout_base64`, but the adapter only executes normalization for `go_version_stdout_base64` at line 245, then falls through to non-empty expected-field checks. Build-source link, special-file, and mutation cases similarly fall through at lines 215–220. Drive all 10 build-source and 12 toolchain cases through the relevant CocoaSkills validators or independently derived fixtures, with platform-specific skips limited to genuinely unsupported operations.

## Independent evidence and passing gates

- Exact head/base and signature verified; worktree clean.
- Exact curator-spec commit `432eb2ee1fe2d6b271e37269f867c8851c325539`; manifest `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`; release record matches, keeps `committed_release_pin_advanced=false`, and emits no claims.
- Candidate tree at CI pin `0c81c1f8d5321d822be2a2817b05aea03e656e15` is byte-identical to `432eb2e` for `conformance` and the rc.6 release record; no `.github` pin change.
- Exact-root focused command passed: 643 passed.
- `python -m mypy` passed: 68 source files.
- Hosted run 30686258237 passed all 14 jobs: Python 3.11–3.14 on Ubuntu, macOS, and Windows, strict mypy, and build artifacts.

Green gates confirm the branch is stable, but the mutation probes prove those gates do not enforce several explicit acceptance criteria. A new reviewer cycle is required after the mappings become behavior-sensitive.