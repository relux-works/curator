# TASK-260720-z2z795 — independent review verdict

## Verdict

**Changes requested → `to-dev`.**

This is ordinary implementation rework, not a stop-the-line boundary. No
external input or human-only architecture decision is needed.

Review run: `RUN-260729-2ed7a8`  
Accepted base reviewed: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`  
Worktree:
`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

The base/provenance handoff is sound: the canonical clean `main`,
`origin/main`, and the task worktree all resolve to the accepted SHA, and
`TASK-260720-1pvfj5` is independently accepted at `done`. The reviewed source
scope is exactly:

- `src/csk/locking.py`
- `src/csk/transactions.py`
- `tests/test_locking.py`
- `tests/test_transactions.py`

## Material findings

### 1. The rollback/install “no-replace” primitive is not atomic

`src/csk/transactions.py:856` checks whether the destination exists and then
uses `Path.rename()` at line 861. On POSIX, `rename()` replaces a destination
that appears after the check. The helper therefore has a classic TOCTOU window
and does not provide the no-replace guarantee its callers rely on.

This affects live-to-backup, staged-to-live, desired-to-rollback, and
backup-to-live transitions. In particular, rollback can pass its current-state
checks and then overwrite a competing successful target that appears before
the rename. That directly violates the acceptance requirements that rollback
restore only known journal state and that one project rollback cannot overwrite
another success.

Deterministic reviewer probe on macOS:

```text
injected_competing_destination=True
destination_after=journal-preimage
source_exists_after=False
```

The probe inserted `other-project-success` at the destination after
`_path_exists(destination)` returned false but immediately before
`Path.rename()`. The implementation silently replaced it with the journal
preimage.

Required rework:

- use a genuinely atomic no-replace operation for every supported platform and
  for both file and directory targets, following the accepted Go behavior
  reference or an equivalently strong design;
- keep the digest/current-state defense adjacent to an atomic mutation whose
  failure preserves the competing destination;
- add deterministic commit and rollback race regressions proving that a
  competing destination is preserved and the transaction remains recoverable.

### 2. Preparation creates live-target parent directories outside the journal

`src/csk/transactions.py:803` calls
`destination.parent.mkdir(parents=True, exist_ok=True)` for the sibling staged
target. If the live target parent did not exist before the transaction,
preparation creates persistent namespace state that is absent from the journal.
Rollback removes the target sidecars, but never removes that newly created
parent.

Reviewer probe:

```text
before_prepare parent_exists=False
after_prepare parent_exists=True entries=['.csk-txn-…-000.desired']
commit_error=TransactionCorruptionError: stale preimage for 10-context/b: ...
after_rollback first_live_exists=False second=foreign-newer
after_rollback parent_exists=True entries=[]
journal_exists=False
```

The journal and target bytes were rolled back, but the transaction left a new
empty persistent directory. This is not an exact reverse rollback and diverges
from the accepted reference, which does not silently create an unjournaled live
parent while staging.

Required rework:

- either require and safely validate an existing staging/live parent, or model
  parent creation/removal as durable transaction-owned state;
- add a regression that begins with an absent parent, forces a later target
  failure, and proves that rollback restores the complete pre-transaction
  namespace with no empty parent or sidecar left behind.

## Independent validation ledger

- Focused pytest:
  `/opt/homebrew/bin/python3.11 -m pytest -p no:cacheprovider tests/test_locking.py tests/test_transactions.py -q`
  → exit `0`, `23 passed in 0.91s`.
- Strict mypy:
  `.../TASK-260720-z2z795/mypy-venv/bin/python -m mypy --cache-dir=/tmp/TASK-260720-z2z795-review-mypy-cache`
  → exit `0`, `Success: no issues found in 56 source files`.
- Source lint:
  `uvx ruff check src/csk/locking.py src/csk/transactions.py`
  → exit `0`, `All checks passed!`.
- Whitespace validation, including the two untracked new files:
  `git diff --check` plus `git diff --no-index --check /dev/null ...`
  → exit `0`.
- The four-path Ruff invocation exits `1` with eight `SIM117` findings in the
  tests. Ruff is not configured as a repository CI gate and some findings
  predate this task, so this is not the verdict basis; however, the next handoff
  should either make the changed-file lint scope green or state the intended
  lint scope accurately.

Tool-readiness notes: `python` is not on `PATH`, so review used the explicit
Python 3.11 executable. The first ad-hoc probe omitted `PYTHONPATH=src` and
failed import-only with exit `1`; it was immediately rerun with the correct
environment and produced the evidence above. Neither probe modified the task
worktree.

## What passed review

The submitted implementation does provide deterministic unsigned UTF-8 target
ordering, durable canonical journal updates, consumer-last backup retention
when caller class keys encode that order, home-scoped journal discovery,
successful interrupted-commit recovery, reverse target iteration, stale
preimage/generation rejection, and unknown-current rollback refusal in the
covered non-racing paths. The focused tests and strict typing are green.

Those strengths do not close the two mutation-boundary defects above.

## Next producer handoff

1. Replace the check-then-rename helper with an atomic cross-platform
   no-replace implementation and cover competing-destination races.
2. Remove or journal the preparation-time parent-directory side effect and add
   the absent-parent rollback regression.
3. Rerun focused pytest, strict `python -m mypy`, source/changed-file lint,
   `git diff --check`, and the relevant full validation.
4. Update the task-scoped developer outcome with exact commands and results,
   then route through a fresh independent reviewer cycle.

The project has no callable `logbook` CLI/connector in this run; these findings
are therefore persisted in this task-scoped outcome and the task notes.
