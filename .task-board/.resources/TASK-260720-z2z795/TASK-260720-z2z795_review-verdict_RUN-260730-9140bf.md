# TASK-260720-z2z795 independent review verdict

- Run: `RUN-260730-9140bf`
- Role: reviewer
- Date: `2026-07-30`
- Verdict branch: **accepted**
- Goal-bound check: none; this run is not goal-bound
- Commit acknowledgement: not supplied (reviewer-archetype run)
- Source modification by reviewer: none

## Reviewed scope and provenance

The review covered the exact current bytes in:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

The task worktree and remote task ref both point to
`5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`. The instructed base
`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` is an ancestor. The dependency
`TASK-260720-1pvfj5` is `done`. The clean canonical main clone is at
`495ad021847529ce5a544dba415ca2fe19949539`, equal to its local
`origin/main`.

The unstaged scope remains exactly:

- `src/csk/locking.py`
- `src/csk/transactions.py`
- `tests/test_locking.py`
- `tests/test_transactions.py`

Source SHA-256:

```text
88a1d0db58c8eb6834cd460b6067371fce222c010d0b1c2b67cae69670015932  src/csk/locking.py
f479fe784a4dd485d355e41c5344717ad26fbb4eecb10fd52ca9f87d7558b34a  src/csk/transactions.py
a10530435846d48c3eda01d034c3e717bbbc9763857d658970e313c155e0cca2  tests/test_locking.py
3d1a03d559b20de0113a434db56b3c184727b390993d894f9b4811875c1223a8  tests/test_transactions.py
```

## Review result

No acceptance-blocking finding remains.

The four findings from
`TASK-260720-z2z795_rereview-verdict_RUN-260730-25e35f.md` are closed:

1. Every public transaction operation binds the held witness to the engine's
   canonical manager-home identity before state access and rechecks it at
   mutation boundaries.
2. Stable entry targets preserve the final symlink entry, stage and digest its
   destination string, replace/restore/remove the entry itself, and leave
   referents unchanged.
3. Cleanup namespaces whose durable state is `removed` reject and preserve
   either canonical-sidecar or tomb reappearance, including exact-digest
   bytes.
4. Home, project, and build locks create first-use home state and then
   recanonicalize before selecting process ownership and physical lock paths.

The implementation also retains deterministic unsigned-UTF-8 project and
target ordering, build-before-home lock enforcement, manager-home serialization,
durable preparation and cleanup ownership, consumer-last commit, reverse
rollback, stale-preimage defense, cross-project recovery, and concurrent
consumer preservation.

## Independent validation

- Reworked boundary selection: `19 passed, 3 skipped`.
- Prescribed prior lock-integrity selection: `14 passed`.
- Contract-targeted selection: `13 passed`.
- Focused locking/transaction pytest: `124 passed, 4 skipped`.
- Full pytest: `694 passed, 23 skipped`.
- Strict mypy: no issues in 57 source files.
- Ruff lint: passed.
- Ruff format check: passed.
- Package sdist and wheel build: passed.
- `git diff --check`: passed.
- Additional canonical-sidecar reappearance probe: foreign bytes, journal
  tomb, and committed live target were all preserved.

Exact commands, results, hashes, and tool versions are recorded in
`TASK-260720-z2z795_review-validation_RUN-260730-9140bf.log`.

## Platform boundary and handoff

This independent run executed on macOS. One existing real-Windows durability
test and three current real-Windows per-directory case-sensitive first-use
contention cases were collected but skipped. Pure Windows routing and
recanonicalization regressions ran locally; no current-byte real-Windows result
is claimed.

The accepted bytes remain unstaged and uncommitted as required. The reviewer
did not supply `commit_ack`. The commit-owning mover can use the hashes above
to commit exactly this four-file scope.
