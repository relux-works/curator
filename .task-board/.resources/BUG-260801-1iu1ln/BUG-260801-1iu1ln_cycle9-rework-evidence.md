# BUG-260801-1iu1ln cycle-9 developer evidence

## Provenance

- Worktree: /Users/iv/Developer/Wildberries/cocoaskills/.temp/BUG-260801-1iu1ln/worktree
- Branch: task/BUG-260801-1iu1ln-lifecycle-observed-traces
- Exact signed merge base: ba250bfc4dfe104a160eadd5b5f4e340693bf892
- Parent: 120be14d31e02ad6c734a3f1a3659d05880933cd
- Signed commit: dbd23cbf24602ccf5125ca78510d9eeb002d0bae
- HEAD and base git verify-commit: exit 0; clean worktree; no tag at HEAD

## Repair

The publication binding now witnesses the destination immediately after the underlying POSIX renameat2/renameatx_np or Windows MoveFileExW primitive. The staged tree must match that raw handoff with only the root rename ctime transition allowed, and the raw handoff must exactly match the state returned by a wrapping CocoaSkills seam. This detects captured root fchmod/restore and transient live-name rename-away/restore without misclassifying the later legitimate cache-root seal. Two exact regressions cover captured POSIX descriptor-relative callables and equivalent supported Windows path rename semantics.

## Direct gates

- Two new reviewer regressions: exit 0, 2 passed
- Cycle-8 plus cycle-9 regression barrier: exit 0, 6 passed
- Unsabotaged canonical lifecycle cases: exit 0, 32 passed
- Canonical/scalar/classification/helper gate: exit 0, 417 passed; all 378 scalar mutations rejected
- Full authenticated exact-root protocol module: exit 0, 869 passed
- Focused product regressions: exit 0, 3 passed
- Install/global/currentness/installer-transaction suites: exit 0, 131 passed
- Transaction/GC/status suites: exit 0, 111 passed and 1 expected platform skip
- Strict mypy: exit 0, no issues in 68 source files
- compileall: exit 0
- Uncommitted and committed diff checks: exit 0
- Cycle-parent product/release guard and exact-base pyproject/CI/changelog guard: exit 0
- Isolated PEP 517 build: exit 0; sdist and wheel version 0.12.6.dev46+gdbd23cbf2
- Twine and sdist membership checks: exit 0
- HEAD/base signatures, exact merge base, clean tree and no-tag checks: exit 0

No PR, main, tag, release, claim, pin, schema-v7, CI, changelog, pyproject or product-source change was made in cycle 9.