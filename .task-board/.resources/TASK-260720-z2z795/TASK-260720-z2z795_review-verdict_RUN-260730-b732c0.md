# TASK-260720-z2z795 independent reviewer verdict

- Run: `RUN-260730-b732c0`
- Role: reviewer
- Date: `2026-07-30`
- Verdict branch: **accepted**
- Route: `done`
- Goal-bound check: none; this run is not goal-bound
- Operator directive: reviewed and acknowledged
- Commit acknowledgement: not supplied
- Source modification by reviewer: none

## Reviewed state

The review covered the exact current unstaged bytes in:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

The signed task HEAD is
`a5ec015c0b5b51149c81a679ce259920a3523374`. The accepted handoff base
`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor. The
prerequisite `TASK-260720-1pvfj5` is `done`.

Current SHA-256:

```text
4a743347eefb084f1051dfbf5d1528eaf64c0e5188aaa9aafcf13461352433c9  src/csk/locking.py
6f5712bf7d0cf8cafdabc7fe5811e918b20ed7c564eb27b23591880dcf5cb6b2  src/csk/transactions.py
e354cafcee95ba3891638651ecbe3ef40d10a76a2c41170ffbc7020284c04d29  tests/test_locking.py
ea56a97b97b3fddb34fc47099728c45f930117922fb709c454a64a16bc5ecb65  tests/test_transactions.py
```

The current delta is exactly `src/csk/locking.py`, `tests/test_locking.py`,
and the previously accepted `tests/test_transactions.py` patch. Nothing is
staged. The Windows test-only patch remains byte-identical to the input reviewed
by `RUN-260730-0c7fda`.

## Review conclusion

No acceptance-blocking finding remains.

The rework closes every finding from `RUN-260730-0c7fda`:

1. Darwin stored-path canonicalization uses the kernel `F_GETPATH` witness on
   an `O_EVTONLY`/no-follow descriptor and no longer enumerates every parent
   directory for every contender.
2. The acquisition deadline is established before manager-home preparation,
   recanonicalization, and lock-directory creation, so preparation consumes the
   same timeout budget as OS-lock contention.
3. Deterministic coverage synchronizes twelve contenders before preparation,
   rejects pre-lock parent scans, and proves all consumers complete.
4. Twenty independent Python/pytest processes completed the prior flaky
   manager-home handoff regression.
5. The accepted Windows permission-stimulus patch is preserved exactly.

The implementation remains aligned with the normative manager profile:
project locks retain canonical unsigned-UTF-8 ordering, build locks precede and
are released before the manager-home lock, shared mutation remains serialized
under the manager-home witness, and the already accepted transaction engine
retains deterministic consumer-last commit, reverse rollback, stale-preimage
defense, and cross-project recovery behavior.

## Independent validation

- repaired boundary: `4 passed`;
- fresh-process handoff stress: `20/20` passed;
- prior reviewer regressions: `14 passed`;
- transaction contract selection: `13 passed`;
- focused locking/transaction pytest: `132 passed, 4 skipped`;
- full Python 3.14 pytest: `811 passed, 58 skipped`;
- strict mypy 2.1 and 2.3: clean in `62 source files`;
- Ruff lint and format: clean;
- package sdist and wheel: built successfully;
- wheel `locking.py` and `transactions.py` hashes match the reviewed sources;
- signed HEAD, accepted-base ancestry, diff checks, and nothing-staged checks:
  green.

Exact commands, outputs, hashes, provenance, and platform boundaries are
recorded in
`TASK-260720-z2z795_review-validation_RUN-260730-b732c0.log`.

## Publication boundary

The reviewer did not modify source, stage, commit, push, tag, release, or
supply `commit_ack`. The current patch remains unstaged and uncommitted. The
commit-owning mover should commit only the reviewed hashes and run the real
Windows Python 3.11-3.14 matrix against the committed bytes before publication.

