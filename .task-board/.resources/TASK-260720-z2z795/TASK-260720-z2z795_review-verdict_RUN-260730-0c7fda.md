# TASK-260720-z2z795 independent reviewer verdict

- Run: `RUN-260730-0c7fda`
- Role: reviewer
- Date: `2026-07-30`
- Verdict branch: **changes requested**
- Route: `to-dev`
- Goal-bound check: none; this run is not goal-bound
- Operator directives: none
- Commit acknowledgement: not supplied
- Source modification by reviewer: none

## Reviewed state

The review covered the exact current bytes in:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

The signed task HEAD and its local upstream are:

`a5ec015c0b5b51149c81a679ce259920a3523374`

`origin/main` is `495ad021847529ce5a544dba415ca2fe19949539` and is an
ancestor of HEAD. The accepted handoff base
`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` is also an ancestor.

The current unstaged delta is exactly `tests/test_transactions.py`. Product
code remains byte-identical to signed HEAD.

Current SHA-256:

```text
6ca6f7b792fba0abd5e25109975e1928b403e5ceb818a45cf48545fc6c5a6aa6  src/csk/locking.py
6f5712bf7d0cf8cafdabc7fe5811e918b20ed7c564eb27b23591880dcf5cb6b2  src/csk/transactions.py
5b8e7a1a91a3e442e905c3be0037ecdb5747c1b1d2cffd464246af75df80d582  tests/test_locking.py
ea56a97b97b3fddb34fc47099728c45f930117922fb709c454a64a16bc5ecb65  tests/test_transactions.py
```

## Acceptance-blocking finding

`ManagerHomeLock` does not reliably complete concurrent-consumer handoffs on
macOS. The task-owned regression
`test_repeated_concurrent_home_consumers_survive_release_handoffs` failed in
the independent full suite and again on the fourth fresh-process stress
iteration because at least one worker remained alive after the seven-second
join boundary.

A six-second faulthandler capture located the stampede before filesystem lock
acquisition: concurrent workers were still in
`_canonicalize_darwin_existing_path()` at `entry.stat(follow_symlinks=False)`
while resolving the same manager home. `_ExclusiveFileLock.__enter__()` calls
`_prepare_for_acquire()` before establishing `start = time.monotonic()`, so
the configured five-second timeout does not bound this canonicalization work.
The result is a nondeterministic task-scoped concurrency failure, not an
external blocker and not an unrelated suite failure.

Required rework:

1. Bound or eliminate the per-contender macOS stored-spelling scan for the
   same manager-home identity while preserving physical-alias and first-use
   correctness.
2. Make lock acquisition's timeout/deadline account for preparation and
   canonicalization, or otherwise prove a bounded total acquisition call.
3. Add deterministic concurrency coverage that controls the preparation
   interleaving and proves all same-home consumers finish without a
   canonicalization stampede. Retain a repeated fresh-process stress run.
4. Re-run the focused suite and full Python 3.14 pytest until both are green,
   along with strict mypy, Ruff, build, diff, and the prior lock/contract
   regressions.

## Other review evidence

The current Windows permission-stimulus rework is sound:

- repaired regressions: `2 passed`;
- forced Windows one-bit identity simulation: `2 passed`;
- negative control with permission validation disabled: both tests failed as
  expected, proving the tests are sensitive to the intended rejection;
- prior lock-integrity regressions: `14 passed`;
- contract-targeted regressions: `13 passed`;
- focused locking/transaction pytest: `129 passed, 4 skipped`;
- strict mypy 2.1 and 2.3: clean in `62 source files`;
- Ruff lint and format: clean;
- package sdist and wheel build: passed;
- wheel `locking.py` and `transactions.py` hashes match the reviewed sources;
- full Python 3.14 pytest: **failed**, `1 failed, 807 passed, 58 skipped`.

Exact commands, outputs, hashes, reproduction details, and raw-log names are
recorded in
`TASK-260720-z2z795_review-validation_RUN-260730-0c7fda.log`.

