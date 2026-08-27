# BUG-260801-1iu1ln cycle-3 reviewer verdict

Verdict: changes requested. Route: to-dev.

## Scope and provenance

Reviewed the dedicated CocoaSkills worktree at `/Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree`, branch `task/BUG-260801-1iu1ln-lifecycle-observed-traces`, signed tip `bb2e5801d3f4c31e48018028097b525238126b33`. Its parent is signed `afc385f6cb24e12b9ff3ac83bc6d1036f3ea3eef`; the preserved earlier commit is signed `9362cc8c076a85a49c04c82e76026d6f7473a311`; merge-base is the exact required signed base `ba250bfc4dfe104a160eadd5b5f4e340693bf892`. All four commits pass `git verify-commit` with the expected ECDSA signer. The source worktree is clean and HEAD has no tag. This reviewer run is not goal-bound.

The three cycle-3 repairs work as submitted: all seven existing product-seam sabotage tests pass, including staged private-artifact execution, guardless GC lock ownership, and first-journal-only recovery. Acceptance still fails because adjacent normative fields remain insensitive to actual CocoaSkills behavior.

## Required findings

### 1. GC artifact execution is invisible when the artifact is not argv element zero

At `tests/protocol_lifecycle_observations.py:1276`, the GC observer examines only `command[0]`; `artifact_executed` is then derived from that partial list at line 1510. A sabotage wrapper around the real POSIX cache collection executed a protected cache artifact as `/bin/sh <protected-artifact>` during each GC collection. Six executions completed with return code 0. The observer still returned `artifact_executed=false`, and the complete `locked-mark-and-sweep-compiled-cache` object remained equal to the normative vector.

This is a direct product-seam false negative. The neighboring private-build observer correctly checks every argv element, but the GC observer does not.

### 2. Recovery project identity is reduced to a basename proxy

The multi-journal predicate at lines 2457-2466 requires only that project identities be distinct. The normative `journal_owner` is then produced with `Path(raw["project_identity"]).name` at line 2516.

A sabotage rewrote the exact owner of `transaction-global-17` from the fixture canonical global identity to a different `wrong-owner/global` path with the same basename. Both journals still resumed and were removed, the observer still reported `journal_owner="global"`, and the complete recovery object remained equal to the normative vector. Exact transaction-to-project identity is therefore not bound, despite the acceptance criterion explicitly requiring recovery transaction/project identity.

### 3. In-place GC entry mutation is not observed by the entry-adoption field

At lines 1435-1440, `entry_adopted` checks only whether a new directory name appeared. It does not compare content, modes, or the complete protected-entry state.

A sabotage performed two real in-place permission repairs on the rejected protected entry before collection. The complete GC case still equaled the vector with `entry_adopted=false`. This permits an in-place chmod/adoption path to survive the normative comparison.

## Required rework

- Trace GC artifact execution across every argv element and the supported process-launch seams; add a sabotage test that executes a protected artifact as a later argv element and requires complete case inequality.
- Bind every recovery transaction ID to its exact expected canonical project identity, not basename or uniqueness. Also tie the reported recovery cache/project identity fields to journal, marker, or recovered state. Add the same-basename wrong-owner sabotage.
- Replace the directory-name proxy for `entry_adopted` with complete protected-entry state/mutation evidence and an explicit in-place permission-repair/adoption trap.
- Preserve all seven existing sabotage tests and the cycle-3 per-call lock and multi-journal fixes.
- Correct the LOGBOOK assertion that all named GC and recovery fields are individually backed, then rerun the full gate set in a new signed commit.

No external blocker or human-only decision exists. This is ordinary implementation rework.

## Reviewer validation

- Authenticated candidate root manifest SHA-256: `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`.
- Seven submitted sabotage probes: 7 passed in 287.17s.
- Canonical plus scalar node sets: 410 passed in 48.75s, comprising 32 lifecycle cases and 378 exhaustive scalar-leaf mutations.
- Full authenticated exact-root conformance: 838 passed in 379.44s.
- Focused product regressions: 3 passed in 4.10s.
- Installer/global/currentness/installer-transaction suite: 131 passed in 142.65s.
- Transaction/GC/status suite: 111 passed and 1 expected platform skip in 18.73s.
- Strict configured mypy: no issues in 68 source files.
- `compileall`, exact-base `git diff --check`, forbidden release-surface/pin diff checks, signature verification, and no-tag check: exit 0.
- Isolated PEP 517 build produced sdist and wheel `0.12.6.dev40+gbb2e5801d`; twine passed both and the sdist contains `tests/protocol_lifecycle_observations.py`.
- No pin, schema-v7, tag, release, claim, PR, main, CI, changelog, or pyproject change was made.