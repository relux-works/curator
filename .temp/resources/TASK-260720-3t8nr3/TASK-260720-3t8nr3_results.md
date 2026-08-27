# TASK-260720-3t8nr3 implementation and validation evidence

## Workspace

- Repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-3t8nr3/worktree`
- Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid`
- Recorded base SHA: `b3a5031ed551b27a298eef486a068b5175beaacc`
- The task worktree was created only after the canonical clean `main` clone
  fast-forwarded to signed `origin/main` and both dependency handoffs were
  verified. The implementation remains uncommitted for reviewer inspection.

## Delivered behavior

- Real project and applicable hybrid installs now plan active build providers,
  run `go-v1` misses provider-first and command-lexically in one
  operation-private toolchain staging area outside the manager-home lock, and
  publish only verified cache winners under that lock.
- A project lock spans the operation. Under the manager-home lock, the
  installer recovers durable transactions; revalidates generations, target
  preimages, closure refs, provider snapshots, collisions, and cache winners;
  then prepares and commits one durable target plan.
- The target plan contains granular project/hybrid context and marker entries,
  script runtime entries, compiled and script shims, environment files,
  adapter mirrors and ledgers, stale removals, and the consumer registry.
  Transaction class ordering makes the consumer target last; transaction
  rollback restores committed targets in reverse order while the lock remains
  held.
- New or changed installs use typed marker v2 with build source and receipt
  identities. Build roots are excluded from context/runtime materialization,
  and compiled shims resolve immutable protected-cache artifacts.
- Adapter symlink/copy modes use per-entry transaction targets plus a later
  ledger target. Aggregate byte-tree targets were rejected because the
  transaction protocol intentionally rejects symlink descendants.
- CLI project install/upgrade no longer takes the legacy outer home lock;
  installer-owned project/build/home lock ordering keeps compilation outside
  the home lock. Audit/registry generation baselines permit only gate-owned
  writes before being rebased. Dead legacy install temporaries are journaled
  stale-removal targets.
- Added seven integration vectors covering audit/build/commit ordering, marker
  v2 and build-root isolation, build and publication failure preservation,
  stale-generation restart, hybrid compiled activation, two-project shared
  cache success, and second-project consumer-last rollback isolation.

## Validation evidence

Every command below ran directly as a standalone process without `tee`.

- Expected-red task suite:
  `python -m pytest -q tests/test_installer_transactions.py` — exit 1,
  4 failed before implementation, as expected.
- Syntax validation:
  `python -m py_compile src/csk/installer.py src/csk/adapters.py
  src/csk/consumers.py tests/test_installer_transactions.py` — exit 0.
- Focused project/hybrid/adapter/closure suite:
  `python -m pytest -q tests/test_install.py tests/test_hybrid_scope.py
  tests/test_adapters.py tests/test_closure_install.py
  tests/test_installer_transactions.py` — exit 0, 84 passed in 804.89s.
- Transaction/locking/cache/planner/activation suite:
  `python -m pytest -q tests/test_transactions.py tests/test_locking.py
  tests/test_build_cache_posix.py tests/test_build_planner.py
  tests/test_build_activation.py` — exit 0, 225 passed and 9 skipped in
  10.07s.
- First strict typing run:
  `python -m mypy` — exit 1, 6 issues: four implementation typing issues and
  two missing generated-version issues. The implementation issues were fixed;
  the documented package build generated the ignored SCM version module.
- First broad suite:
  `python -m pytest -q` — exit 1, 24 failed, 1103 passed, 98 skipped in
  1311.70s. It exposed the legacy CLI outer-home-lock conflict, gate-owned
  audit generation writes, orphan cleanup, and a missing unknown-agent
  warning. All four causes were corrected.
- Corrected affected-module suite:
  `python -m pytest -q tests/test_audit_cli.py tests/test_audit_registry.py
  tests/test_cli.py tests/test_e2e.py tests/test_gc.py
  tests/test_global_install.py::test_project_command_shim_shadows_global_command
  tests/test_global_install.py::test_runtime_gc_keeps_global_only_runtime` —
  exit 0, 124 passed and 1 skipped in 206.99s.
- Decisive full suite:
  `python -m pytest -q` — exit 0, 1127 passed and 98 skipped in 1301.10s.
- Final task-specific suite:
  `python -m pytest -q tests/test_installer_transactions.py` — exit 0,
  7 passed in 80.20s.
- Lint:
  `uvx ruff check src/csk/installer.py src/csk/adapters.py src/csk/cli.py
  src/csk/consumers.py tests/test_installer_transactions.py
  tests/test_install.py` — exit 0.
- A whole-file optional formatter probe on legacy touched files:
  `uvx ruff format --check ...` — exit 1 because the repository's existing
  `installer.py`, `adapters.py`, `cli.py`, and `test_install.py` formatting is
  not Ruff-format canonical. No thousands-line style-only rewrite was made.
  The wholly new test and standalone consumer module were formatted, and
  `uvx ruff format --check tests/test_installer_transactions.py
  src/csk/consumers.py` exits 0.
- Final strict typing:
  `python -m mypy` — exit 0, no issues in 67 source files.
- Final package build:
  `python -m build` — exit 0; sdist and wheel built successfully.
- Distribution validation:
  `python -m twine check dist/*` — exit 0; wheel and sdist passed.
- Patch hygiene:
  `git diff --check` — exit 0.

## Review notes

- No product/API forced fit was required. The adapter representation follows
  the accepted transaction `entry` target model rather than weakening byte
  tree validation.
- A task-board spawn status check attempted automatic lifecycle recovery and
  reported an unrelated acceptance-intent digest mismatch for another run;
  the assigned run itself had no operator directives. This did not affect the
  code worktree or validation.
