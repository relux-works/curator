# TASK-260720-2x6mjn — implementation and validation evidence

## Provenance

- Clean, fast-forwarded `origin/main`: `11160f642d65a8daf3fbcca5401dca5ec80440f9`
- Dependency handoff/base commit: `82d1cfc769d5c056e16f0c120ec3b11e2ccc8dae`
- Task branch: `task/TASK-260720-2x6mjn-pure-build-planner`
- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`
- No files were staged, committed, or pushed.

## Implemented scope

- Added a typed, side-effect-free compiled-build planner in `src/csk/builds/planner.py`.
- Produced provider-first and command-lexical plans with exact input, cache key, target, artifact, cache-hit flag, and the five required outcomes:
  `cache-hit`, `would-preflight-and-build`, `would-rebuild-untrusted-cache`, `corrupt`, and `unsupported`.
- Routed project and global dry-run planning through validation, frozen-source, closure/collision, source-audit, registry/attestation, moved-tag, MCP, and system-requirement gates before package-independent Go probes and protected cache reads.
- Added read-only audit-registry and HTTP-fetch paths that do not create or refresh persistent state.
- Made project/global dry-run CLI paths bypass mutation-lock construction and recovery.
- Added bounded full-plan generation rechecks that retry once and then report `concurrent_state_change`.
- Kept compilation, cache publication, markers, shims, adapters, target swaps, and garbage collection outside this task, as assigned to downstream work.

## Acceptance evidence

- Gate ordering: project and global integration tests assert all validation/trust gates precede build planning.
- Plan determinism: planner tests assert provider-first ordering, lexical command ordering, exact build inputs/keys/targets, and all required cache outcomes.
- Source-aware Go purity: planner tests forbid `go list`/`go build`; only package-independent toolchain probes are admitted.
- Persistent-state purity: before/after tests cover audit records, registry/cache state, fingerprints, Go cache, staging/cache entries, journal/runtime/context, markers, shims/adapters/consumers, locks, and GC-visible state.
- Concurrency: project and global tests verify whole-plan retry and repeated-change reporting.
- Read-only registry behavior: tests verify missing and existing registry state remain byte-for-byte unchanged and HTTP cache state is neither created nor refreshed.

## Command evidence

### Expected-red

- Focused planner/purity test selection before implementation: exit `4`.
  Expected rationale: pytest collection failed because `csk.builds.planner` did not yet exist.

### Final task-scoped gates

- Focused pytest selection covering planner, gate order, dry-run purity, read-only registry, CLI lock routing, and generation retries: exit `0`; `20 passed, 1 skipped in 2.65s`.
  The skip is the platform-specific `O_NOATIME` assertion because macOS does not expose `os.O_NOATIME`; the read-only state tests ran green.
- `python -m mypy`: exit `0`; `Success: no issues found in 67 source files`.
- `python -m compileall -q src/csk tests`: exit `0`.
- `python -m tabnanny src/csk tests`: exit `0`.
- `git diff --check`: exit `0`.
- `python -m build`: exit `0`; sdist and wheel built successfully.
- `python -m twine check dist/*`: exit `0`; both artifacts passed.
- `git diff --cached --quiet`: exit `0`; no staged changes.

### Compatibility suites

- `tests/test_audit_registry.py`: exit `0`; `46 passed`.
- `tests/test_install.py`: exit `0`; `47 passed`.
- `tests/test_global_install.py`: exit `0`; `27 passed`.
- `tests/test_cli.py`: exit `0`; `53 passed`.

Ruff is not installed in the project environment and the repository has no Ruff configuration. The repository's available Python lint/validation gates (`compileall`, `tabnanny`, strict mypy, and diff whitespace validation) all exited `0`.

## Review cycle 2 — requested changes resolved

Independent review confirmed the pure planner and dry-run contract, then found
two regressions in the unchanged non-dry-run global-install path plus missing
coverage for the newly shared real-install gates.

### Rework

- Restored per-declaration closure error isolation. The canonical combined
  closure remains the fast path; only a failed combined resolution is
  decomposed to identify unavailable declarations, then the surviving
  declarations are rebuilt as one canonical provider-first closure.
- Prevented partial global installs from narrowing the cleanup keep-set. Any
  resolution or dependency error now suppresses cleanup, preserving existing
  installed skills and shims.
- Skipped toolchain/cache build planning whenever a global install is partial,
  while continuing to install healthy skills. Dry-run generation is still
  rechecked without entering the toolchain/cache planner.
- Added explicit coverage for shared-provider closure order, cross-declaration
  closure conflicts, real-install MCP/registry ordering and failure, healthy
  build-skill installation during unrelated dependency failure, and
  preservation of an existing skill during unrelated resolution failure.

### Cycle-2 command evidence

- Reviewer regression probes before rework: exit `1`; `2 failed` as expected,
  reproducing the data-loss and healthy-skill suppression bugs.
- Reviewer regression probes after rework: exit `0`; `2 passed`.
- Focused repository rework selection: exit `0`; `6 passed`.
- `tests/test_global_install.py`: exit `0`; `35 passed`.
- Task-focused pytest (`test_build_planner.py`, `test_audit_registry.py`,
  `test_cli.py`, `test_global_install.py`, `test_install.py`): exit `0`;
  `189 passed, 1 skipped in 55.66s`.
- Full pytest on the current tree: exit `0`;
  `1016 passed, 92 skipped in 99.04s`.
- `python -m mypy`: exit `0`;
  `Success: no issues found in 67 source files`.
- `python -m compileall -q src/csk`: exit `0`.
- `python -m tabnanny src/csk tests`: exit `0`.
- `git diff --check`: exit `0`.
- `python -m build`: exit `0`; sdist and wheel built successfully.
- `python -m twine check dist/*`: exit `0`; both artifacts passed.
- `git diff --cached --quiet`: exit `0`; no staged changes.

No files were staged, committed, or pushed during rework.
