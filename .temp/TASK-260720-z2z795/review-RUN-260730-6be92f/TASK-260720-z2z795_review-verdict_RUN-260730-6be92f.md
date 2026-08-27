# TASK-260720-z2z795 independent reviewer verdict

- Run: `RUN-260730-6be92f`
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

The task worktree and remote task ref both point to
`108ae51e549566b0b7e1ff4a8a2424e8cb4fcc77`. The instructed accepted base
`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` is an ancestor.

The current unstaged delta is exactly:

- `src/csk/locking.py`
- `tests/test_locking.py`
- `tests/test_transactions.py`

`src/csk/transactions.py` is unchanged from the previously accepted review and
retains SHA-256
`f479fe784a4dd485d355e41c5344717ad26fbb4eecb10fd52ca9f87d7558b34a`.

Current source SHA-256:

```text
6ca6f7b792fba0abd5e25109975e1928b403e5ceb818a45cf48545fc6c5a6aa6  src/csk/locking.py
f479fe784a4dd485d355e41c5344717ad26fbb4eecb10fd52ca9f87d7558b34a  src/csk/transactions.py
5b8e7a1a91a3e442e905c3be0037ecdb5747c1b1d2cffd464246af75df80d582  tests/test_locking.py
3f4edc3fadff19974c0194a9b9ad3535c5b4032308158626f76339170ac937af  tests/test_transactions.py
```

## Review result

No acceptance-blocking finding remains.

The Windows failure at the committed input was correctly traced to opening the
stable lock descriptor in CRT text mode. The implementation now ORs the
platform-optional `O_BINARY` flag into the one shared stable-descriptor open
path. This covers manager-home, project, and build locks without changing POSIX
behavior.

The accompanying test changes preserve the adversarial assertions while
respecting Windows byte-range lock semantics:

- stable descriptors must carry `O_BINARY`;
- held publication bytes are inspected through the owning descriptor;
- Windows verifies that the live stable handle prevents path rename, while the
  existing POSIX replacement-race attack remains active;
- witness-loss injection writes through the owning descriptor and is still
  detected at the transaction mutation boundary.

Official contracts support both parts of the fix: Python requires `O_BINARY`
for binary `os.open` on Windows; the Microsoft CRT suppresses CRLF translation
in binary mode; and `LockFileEx` denies access to the locked range through a
second handle opened by the same process.

The previously accepted transaction engine remains intact: canonical
unsigned-UTF-8 project ordering, build-before-home release, no outer lock
acquisition under the home lock, durable canonical journals, deterministic
target order, consumer-last durability, reverse rollback, desired-digest
rollback defense, cross-project recovery, stale-preimage/generation rejection,
and concurrent consumer preservation.

## Independent validation

- Prior lock-integrity selection: `14 passed`.
- Contract-targeted transaction selection: `13 passed`.
- Windows-sensitive selection on Python 3.11: `21 passed`.
- Windows-sensitive selection on Python 3.14: `21 passed`.
- Focused locking/transaction pytest: `125 passed, 4 skipped`.
- Full pytest: `804 passed, 58 skipped`.
- Strict mypy: no issues in `62 source files`.
- Ruff lint: passed.
- Ruff format check: passed.
- Package sdist and wheel build: passed.
- `git diff --check`: passed.
- Independent all-lock-class descriptor probe: home/project/build each opened
  with the binary flag.

The first descriptor probe already reported the correct three-class assertion
but its test harness left `os.open` patched during temporary-directory cleanup
and exited nonzero on `dir_fd`. The corrected probe restored `os.open` before
teardown and exited zero. This was a reviewer-probe defect, not a product
failure.

Exact commands, results, hashes, and tool versions are recorded in
`TASK-260720-z2z795_review-validation_RUN-260730-6be92f.log`.

## Platform boundary and commit-owner handoff

This independent run executed on macOS. The focused suite's four skips are the
one real-Windows durability test and three real-Windows per-directory
case-sensitive first-use contention cases. The current three-file patch remains
unstaged and uncommitted under the task constraint, so no current-byte
real-Windows CI result is claimed.

The accepted verdict is for the exact hashes above. The commit-owning mover
should commit only that scope and rerun the Windows Python
3.11/3.12/3.13/3.14 matrix before publication. The reviewer supplied no
`commit_ack`.
