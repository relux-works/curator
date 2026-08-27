# BUG-260801-1iu1ln cycle-8 developer evidence

## Provenance

- Worktree: /Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree
- Branch: task/BUG-260801-1iu1ln-lifecycle-observed-traces
- Exact signed merge base: ba250bfc4dfe104a160eadd5b5f4e340693bf892
- Cycle-8 parent: a0046fdfbd37ecce4c5d6d0e21152628c2d2432f
- Signed commit: 120be14d31e02ad6c734a3f1a3659d05880933cd
- git verify-commit HEAD: exit 0, good ECDSA signature for oparin@me.com
- Clean worktree, exact merge base, committed diff hygiene, restricted release-surface guard, and no-tag guard: exit 0

## Repair

Cycle-8 replaces alias-specific final-byte assumptions with bounded recursive persistent-state witnesses containing node type, mode, device/inode, link count, ownership, size, bytes or link target, mtime_ns, ctime_ns, flags and file attributes. Atomic publication compares the staged tree to the live tree after the no-replace handoff and permits only the moved root directory ctime change caused by rename. Dry-run, planning, currentness, GC rejection and private-failure observations use the witness. Private-build failure protection now spans the project, config, source and manager persistent surfaces around the exact private-build phase; stable lock records are explicitly classified as coordination state.

Four regressions cover captured descriptor-relative os aliases after publication, io.open write/fsync/restore mutations, the formerly unwatched private-failure Skillfile.json, and exact byte/mode/inode/mtime restoration with persistent ctime evidence.

## Direct gates

- Pre-fix three-regression expected-red gate: exit 1, 3 failed in 204.46s. This is the intended proof that all three cycle-7 survivors reproduced before repair.
- Darwin ctime restoration premise probe: exit 0; bytes, mode, inode and mtime restored while ctime changed.
- Four new regressions: exit 0, 4 passed in 208.37s.
- Unsabotaged canonical lifecycle vectors: exit 0, 32 passed in 68.05s.
- Canonical plus exhaustive scalar/classification/helper audit: exit 0, 417 passed in 66.41s.
- Retained sabotage barrier: exit 0, 28 passed in 1427.99s.
- Full authenticated exact-root protocol module: exit 0, 867 passed in 1494.82s.
- Focused product regressions: exit 0, 3 passed in 3.20s.
- Install/global/currentness/installer-transaction suites: exit 0, 131 passed in 127.84s.
- Transaction/GC/status suites: exit 0, 111 passed and 1 platform skip in 16.68s.
- Strict mypy: exit 0, no issues in 68 source files.
- compileall: exit 0.
- Uncommitted, staged and committed diff checks: exit 0.
- Exact-base release guard and cycle-parent product-source/test_cli guard: exit 0.
- Isolated detached signed-tree PEP 517 build: exit 0; sdist and wheel version 0.12.6.dev45+g120be14d3.
- Twine check for both distributions: exit 0.
- Sdist membership for LOGBOOK.md and both changed lifecycle test files: exit 0.
- Signature, clean-tree, exact-parent, exact-merge-base and no-tag checks: exit 0.

## Additional diagnostic

A broader whole-repository pytest diagnostic exited 1 with 1 failed, 2135 passed and 54 skipped in 1667.40s. The sole failure was tests/test_cli.py::test_cli_lock_contention_returns_lock_exit: it expected EXIT_LOCK 3 but the unchanged product returned EXIT_GENERAL 1 for legacy lock state cannot be migrated online. The exact standalone test reproduced the same result with exit 1 in 1.09s. This cycle has zero diff in src/csk and tests/test_cli.py relative to signed parent a0046fdf, and the required exact-root and related suites above are green.

No PR, main, tag, release, claim, pin, schema-v7, CI, changelog, pyproject or product-source change was made.