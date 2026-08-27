# BUG-260801-1iu1ln cycle-3 rework evidence

## Provenance
- Worktree: /Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree
- Branch: task/BUG-260801-1iu1ln-lifecycle-observed-traces
- Exact signed base and merge-base: ba250bfc4dfe104a160eadd5b5f4e340693bf892
- Preserved signed commits: 9362cc8c076a85a49c04c82e76026d6f7473a311 and afc385f6cb24e12b9ff3ac83bc6d1036f3ea3eef
- Signed cycle-3 commit: bb2e5801d3f4c31e48018028097b525238126b33
- HEAD parent is afc385f6cb24e12b9ff3ac83bc6d1036f3ea3eef; verify-commit for base, parent, and HEAD exited 0 with the expected ECDSA signer.
- Final source worktree is clean; HEAD has no tag.

## Cycle-3 repairs
- The successful all-misses observer records every Popen argv across the operation and derives artifacts_executed from exact operation-private artifact paths. A sabotage wrapper actually executes both staged artifacts with return code 0 and now changes the normative case.
- The GC observer records a separate trace for each of four guardless and one guarded collect_runtime invocation. Every guardless call must acquire, validate, and forward its own ManagerHomeLock witness; the guarded call must validate and forward the supplied witness without another lock; project/build locks and fixture-only transaction locks cannot satisfy only_lock.
- The recovery fixture creates two rolling-back journals with distinct transaction IDs and project identities. One public recover call must resume both exact IDs, restore state, and remove both journals before scan_scope can report all-incomplete-journals.
- The adjacent private-build, GC, and recovery fields were audited for the same proxy/literal pattern. LOGBOOK now accurately records seven independent product-seam sabotage probes.

## Expected-red proof
- Initial three-probe command: exit 1. Private-artifact execution and first-only recovery failed because their sabotaged observations still equaled the normative cases. The initial GC sabotage fixture also failed structurally because its no-op witness lacked home_identity.
- Corrected GC expected-red command: exit 1; the guardless sabotage completed against real GC state and the observed case still equaled the normative vector before repair.
- After implementation, the three new probes: exit 0; 3 passed in 119.51s.

## Direct final gates
- Canonical authenticated lifecycle subset: exit 0; 32 passed in 39.27s.
- Exhaustive scalar-leaf mutation audit: exit 0; 378 passed in 39.17s.
- All seven product-seam sabotage tests: exit 0; 7 passed in 272.73s.
- Full authenticated exact-root conformance: exit 0; 838 passed in 351.12s.
- Preserved focused product regressions: exit 0; 3 passed in 3.10s.
- Installer/global/currentness/transaction product suite: exit 0; 131 passed in 119.14s.
- Transaction/GC/status product suite: exit 0; 111 passed and 1 expected platform skip in 13.66s.
- Strict configured mypy: exit 0; no issues in 68 source files.
- compileall src tests: exit 0.
- Exact-base committed git diff --check: exit 0.
- Forbidden release-surface diff guard for pyproject, CI, changelog, and generated version file: exit 0.
- Signed commit verification: exit 0; good ECDSA signature.
- Isolated clean signed-tree PEP 517 build: exit 0; built sdist and wheel 0.12.6.dev40+gbb2e5801d.
- Twine check of both isolated artifacts: exit 0; PASSED.
- Sdist lifecycle-observer membership check: exit 0.
- Post-build clean-tree, signature, diff, release-surface, and no-tag checks all passed.

No pin, schema-v7, tag, release, claim, PR, main landing, CI, changelog, or pyproject change was made. This signed commit is for later PR19 integration only.