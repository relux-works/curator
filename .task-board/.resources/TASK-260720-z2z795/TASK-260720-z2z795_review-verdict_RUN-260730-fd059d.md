# TASK-260720-z2z795 independent reviewer verdict

- Run: `RUN-260730-fd059d`
- Role: reviewer
- Date: `2026-07-30`
- Verdict branch: **accepted**
- Goal-bound check: none; this run is not goal-bound
- Operator directives: none
- Commit acknowledgement: not supplied (reviewer-archetype run)
- Source modification by reviewer: none

## Reviewed scope and provenance

The review covered the exact current bytes in:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

The signed task HEAD and its local upstream are both:

`d0d91788bc01c6590dec49a0dd296a78e95fd44b`

The instructed accepted base
`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` is an ancestor.

The current unstaged delta is exactly:

- `src/csk/transactions.py`
- `tests/test_transactions.py`

Current source SHA-256:

```text
6ca6f7b792fba0abd5e25109975e1928b403e5ceb818a45cf48545fc6c5a6aa6  src/csk/locking.py
6f5712bf7d0cf8cafdabc7fe5811e918b20ed7c564eb27b23591880dcf5cb6b2  src/csk/transactions.py
5b8e7a1a91a3e442e905c3be0037ecdb5747c1b1d2cffd464246af75df80d582  tests/test_locking.py
1d21d2513b2ef06f435c905eda8a0bf8f3e74c156bd07bea67557119420028a6  tests/test_transactions.py
```

## Review result

No acceptance-blocking finding remains.

The rework correctly closes the two Windows CI failure clusters at signed HEAD
`d0d9178`:

1. Transaction namespace tests now construct forbidden lock targets without
   reopening or digesting a held `LockFileEx` path. The engine itself reaches
   canonical manager-home/project/build namespace rejection and preserves the
   owning lock witnesses.
2. Staging and cleanup validation compare the only permission state Windows
   `chmod` can set: read-only versus writable through `stat.S_IWRITE`. The
   comparison is centralized and covers ordinary staging, active crash
   recovery, cleanup trees, and cleanup entries. POSIX continues to compare
   exact modes.

The `_require_phase` cast is guarded by an explicit runtime string and literal
membership check, so it closes the strict-mypy diagnostic without widening
accepted journal values.

The implementation remains aligned with the normative manager transaction
contract: deterministic unsigned-UTF-8 ordering, build-lock release before the
home lock, no outer lock acquisition under the home lock, durable journals,
consumer-last commit, exact reverse rollback, desired-digest rollback defense,
cross-project recovery, stale-preimage/generation refusal, and concurrent
consumer preservation.

## Independent validation

- Windows CI cluster simulation:
  - Python 3.11: `15 passed`
  - Python 3.14: `15 passed`
- Prescribed prior lock-integrity regressions: `14 passed`
- Contract-targeted regressions: `13 passed`
- Focused locking/transaction pytest: `129 passed, 4 skipped`
- Full Python 3.14 pytest: `808 passed, 58 skipped`
- Strict mypy 2.1: clean in `62 source files`
- Strict mypy 2.3: clean in `62 source files`
- Ruff lint and format checks: passed
- Package sdist and wheel build: passed
- `git diff --check`: passed
- Adversarial permission probe:
  - Windows accepts only non-settable-bit normalization and rejects
    read-only/writable transitions.
  - POSIX rejects group-write and execute-bit changes exactly.
- The built wheel contains `csk/transactions.py` with the exact reviewed source
  SHA-256.

Exact commands, outputs, hashes, and tool versions are recorded in
`TASK-260720-z2z795_review-validation_RUN-260730-fd059d.log`.

## Platform boundary and commit-owner handoff

This independent run executed on macOS. The focused suite's four skips are the
existing real-Windows durability and per-directory-case-sensitive first-use
cases. The current two-file patch remains unstaged and uncommitted, so no
current-byte real-Windows CI success is claimed.

The commit-owning mover should commit only the exact reviewed hashes above and
rerun the Windows Python 3.11, 3.12, 3.13, and 3.14 jobs before publication.
The reviewer supplied no `commit_ack`.
