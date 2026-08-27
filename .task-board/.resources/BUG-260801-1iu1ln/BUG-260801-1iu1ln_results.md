# BUG-260801-1iu1ln lifecycle binding evidence

## Signed source
- Worktree: /Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree
- Branch: task/BUG-260801-1iu1ln-lifecycle-observed-traces
- Required signed base: ba250bfc4dfe104a160eadd5b5f4e340693bf892; git verify-commit exit 0.
- Signed task commit: 9362cc8c076a85a49c04c82e76026d6f7473a311; git verify-commit exit 0, good ECDSA signature.
- HEAD parent and merge base both equal the required base. Worktree clean.

## Outcome
- All 32 manager lifecycle vectors are reconstructed from authenticated-fixture CocoaSkills operations/state across bootstrap, closure/build ordering, POSIX/Windows publication, transactions and cross-project ledger, dry run, GC, launchers, planning gates, private builds, recovery, repair, status/currentness, commit/rollback, and upgrades.
- Adapter compares each complete normative case object to observed output and fails closed on unknown cases, clusters, fields, or nested field additions.
- Exhaustive one-scalar-leaf audit covers all 378 leaves: baseline had 104 survivors; repaired coverage rejects all 378 mutations, leaving zero survivors.
- Observations exposed and fixed three product gaps: recovery ran before private builds in project/global install; runtime build-root exposure was not non-current; global upgrade did not fetch transitive closure. Regression tests cover each behavior.
- No pin, schema-v7, tag, release, claim, CI, or pyproject change.

## Direct validation evidence
- Exact-root conformance: exit 0, 831 passed in 127.10s.
- Canonical lifecycle subset: exit 0, 32 passed.
- Exhaustive scalar mutation subset: exit 0, 378 passed.
- Related installer/global/currentness/transaction suites: exit 0, 144 passed in 219.19s.
- Focused product regressions: exit 0, 3 passed.
- Strict mypy: exit 0, no issues in 68 source files.
- compileall src tests: exit 0. Project has no configured standalone formatter/linter.
- git diff --check required-base..HEAD: exit 0.
- Forbidden release-surface diff guard: exit 0.
- Clean signed-tree python -m build: exit 0; produced sdist and wheel for 0.12.6.dev38+g9362cc8c0 and included the new observation helper in the sdist.
- twine check of both signed-tree distributions: exit 0, PASSED.
