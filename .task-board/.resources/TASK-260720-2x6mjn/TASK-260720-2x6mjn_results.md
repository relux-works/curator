# TASK-260720-2x6mjn — implementation and validation evidence

## Provenance

- Clean, fast-forwarded `origin/main` recorded at task start:
  `11160f642d65a8daf3fbcca5401dca5ec80440f9`.
- Accepted dependency handoff and direct task base:
  `82d1cfc769d5c056e16f0c120ec3b11e2ccc8dae`.
- Task branch: `task/TASK-260720-2x6mjn-pure-build-planner`.
- Task worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`.
- No files were staged, committed, or pushed.

## Implemented scope

- Added the typed, side-effect-free compiled-build planner in
  `src/csk/builds/planner.py`.
- Produces provider-first and command-lexical plans containing exact input,
  cache key, target, artifact, cache-hit state, and the required
  `cache-hit`, `would-preflight-and-build`,
  `would-rebuild-untrusted-cache`, `corrupt`, and `unsupported` outcomes.
- Routes project and global dry-run planning through frozen-source,
  validation, closure/collision, source-audit, registry/attestation,
  moved-tag, MCP, and system-requirement gates before package-independent Go
  probes or protected cache reads.
- Adds read-only audit-registry and HTTP-cache paths, mutation-lock-free CLI
  dry-run routing, and bounded full-plan generation retries with
  `concurrent_state_change`.
- Keeps compilation, publication, markers, shims, adapters, target swaps,
  and garbage collection outside this task.

## Review cycle 3 — C1/C2 rework

Independent review accepted the pure planning core and found two regressions
where planning had leaked into established mutating install behavior.

- C1: global real install materialized the full dependency closure, including
  undeclared context-only providers and their inactive commands. Two inactive
  providers could silently shadow the same `global/bin` command.
- C2: project and global real install invoked `plan_builds`, making those
  existing install paths require Go even though no downstream code consumed
  the plan.

The rework establishes an explicit boundary:

- Dry-run uses the validated full closure and performs build planning.
- Real global install resolves and materializes declarations only, retaining
  per-declaration failure isolation and partial-install cleanup safety.
- Real project and global installs do not freeze build providers or invoke
  toolchain/cache planning. They therefore do not require Go.
- Global MCP and registry checks introduced for build planning are confined
  to dry-run; the established real-install audit gate remains in place.
- Added regression tests for undeclared provider/context/runtime leakage,
  inactive command shadowing, real project/global installs without Go,
  planning-only gate routing, and the earlier partial-install B1/B2 cases.
- Recorded the boundary and downstream ownership decision in `LOGBOOK.md`.

## Command evidence

### Expected-red and invocation evidence

- Initial cycle-3 focused regression selection: exit `1`;
  `5 failed, 1 passed`. The five failures reproduced C1/C2 before production
  changes.
- First focused invocation using worktree-local `.venv/bin/python`: exit
  `127`; the task worktree has no local virtual environment. The exact
  selection was rerun with the verified canonical project interpreter.
- Unchanged full reviewer probe archive after rework: exit `1`;
  `1 failed, 5 passed`. The sole failure is
  `test_probe_global_install_with_skill_missing_skill_md`, explicitly
  documented by the reviewer as a non-finding that also fails on the direct
  base.
- That unchanged non-finding against detached base `82d1cfc`: exit `1`;
  `1 failed, 1 passed`, confirming it is not introduced by this task.

### Green task and repository gates

- Cycle-3 focused regression selection after rework: exit `0`;
  `6 passed`.
- Dry-run ordering/purity plus B1/B2 safety selection: exit `0`;
  `12 passed`.
- C1/C2 reviewer finding probes only: exit `0`; `4 passed`. Observed output
  showed only declared global skills, empty transitive global bins, and
  successful real project/global installs with Go removed from `PATH`.
- `tests/test_install.py tests/test_global_install.py`: exit `0`;
  `89 passed`.
- Task-focused pytest
  (`test_build_planner.py`, `test_audit_registry.py`, `test_cli.py`,
  `test_global_install.py`, `test_install.py`): exit `0`;
  `193 passed, 1 skipped`.
- Full `python -m pytest -q`: exit `0`;
  `1020 passed, 92 skipped in 101.37s`.
- Strict `python -m mypy`: exit `0`;
  `Success: no issues found in 67 source files`.
- `python -m compileall -q src/csk tests`: exit `0`.
- `python -m tabnanny src/csk tests`: exit `0`.
- `git diff --check`: exit `0`.
- `git diff --cached --quiet`: exit `0`; no staged changes.
- `python -m build`: exit `0`; sdist and wheel built successfully.
- `python -m twine check dist/*`: exit `0`; both artifacts passed.

Ruff was not run because it is not installed in the project environment and
the repository has no Ruff configuration. Strict mypy, compileall, tabnanny,
and diff whitespace validation all exited `0`.

## Review cycle 4 — F1 test-only rework

Independent review confirmed the production planner and the cycle-3 C1/C2
fixes, then identified one remaining portability and coverage gap: dry-run
correctly fails closed when no trusted Go executable is available, but neither
project nor global scope pinned that behavior, and an unrelated schema-v6
build-root lifecycle test inherited the contributor machine's Go installation.

This rework changes tests and the task logbook only:

- Project and global integration tests now capture an empty operator search
  path and assert the stable
  `go-v1 go_toolchain_missing: captured operator PATH contains no Go executable`
  error.
- Both tests assert fail-closed whole-plan behavior: failed status, no partial
  build rows or planned summaries, and byte/metadata-equivalent watched
  filesystem state before and after.
- Each missing-Go plan includes an unrelated plain skill, pinning that no
  partial plan is returned when the trusted toolchain gate fails.
- The unrelated schema-v6 build-root lifecycle test now injects a deterministic
  trusted toolchain identity, so it remains focused on context filtering and
  passes on contributor machines without Go.
- `LOGBOOK.md` records the fail-closed blast radius and host-independent test
  boundary. No production behavior changed in this cycle.

### Cycle-4 command evidence

- Expected-red Go-less schema-v6 lifecycle test before rework: exit `1`;
  `1 failed`, with the reproduced `go_toolchain_missing` error.
- Targeted cycle-4 tests: exit `0`; `3 passed in 1.34s`.
- Go-less focused suite (`test_install.py`, `test_global_install.py`,
  `test_cli.py`, `test_build_planner.py`): exit `0`;
  `149 passed in 38.69s`.
- Full `python -m pytest -q`: exit `0`;
  `1022 passed, 92 skipped in 86.27s`.
- Strict `python -m mypy`: exit `0`;
  `Success: no issues found in 67 source files`.
- `python -m compileall -q src tests`: exit `0`.
- `python -m tabnanny -q src tests`: exit `0`.
- `git diff --check`: exit `0`.
- `python -m build`: exit `0`; sdist and wheel built successfully and include
  `src/csk/builds/planner.py`.
- `python -m twine check dist/*`: exit `0`; both artifacts passed.

## Handoff state

Cycle-4 developer rework is ready for independent review. The worktree remains
unstaged and uncommitted as required.
