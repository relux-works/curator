# TASK-260720-z2z795 — install transaction engine developer outcome

## Provenance

- Accepted base SHA: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Branch: `task/TASK-260720-z2z795-transaction-engine`
- Worktree: `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- Prerequisite: `TASK-260720-1pvfj5` independently accepted at board status `done`
- Normative contract: `curator-spec/profiles/manager.md` sections 2.5–2.6
- Accepted behavior reference: `curator/.temp/TASK-260720-1pvfj5/rework/composite/internal/transaction`

## Delivered source

- `src/csk/locking.py`
  - canonical physical project identities
  - unsigned UTF-8 multi-project ordering
  - project, optional one-key build, and manager-home lock hierarchy
  - process-wide refusal of new project/build locks while the home lock is held
  - held home-lock witness required by transaction mutation/recovery
  - backward-compatible `GlobalLock` name
- `src/csk/transactions.py`
  - generic mutable file/directory targets, independent of installer/compiler policy
  - canonical domain-separated target digests and explicit `absent`
  - durable canonical journals with transaction/project identity, ordered classes and identifiers, expected preimage or generation, generation digests, backup/staging/rollback paths, desired digests, and per-target commit state
  - deterministic unsigned UTF-8 class/identifier ordering
  - atomic journal creation and fsync-backed journal/target transitions
  - consumer-last support through caller-owned ordered class keys
  - crash completion from preparing, committing, cleanup, and rolling-back phases
  - exact reverse rollback with desired-digest defense before removing current state
  - recovery across every home journal without a project filter
  - bounded regular-file journal reads plus canonical sidecar namespace validation

## Focused coverage

`tests/test_locking.py` and `tests/test_transactions.py` exercise:

- canonical/deduplicated project order including non-ASCII identities
- reversed project order, build-without-project, second build key, build-to-home, project-under-home, and cross-thread project-under-home failures
- deterministic class/identifier order and consumer-last durability
- backup retention through consumer commit
- crash recovery after backup, after install, after target state, and before cleanup
- reverse rollback observation
- stale preimage and stale generation refusal
- rollback refusal to overwrite unknown current bytes
- corrupted journal sidecar redirection refusal
- concurrent consumers preserving both projects
- a failed second project preserving the first project's consumer/context success
- managed removal targets

## Exact validation ledger

All commands were run directly as standalone processes from the task worktree.

1. `/opt/homebrew/bin/python3.11 -m pytest tests/test_locking.py tests/test_transactions.py -q`
   - exit `0`
   - `23 passed in 0.89s`
2. `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/mypy-venv/bin/python -m mypy`
   - exit `0`
   - `Success: no issues found in 56 source files`
3. `uvx ruff check src/csk/locking.py src/csk/transactions.py`
   - exit `0`
   - `All checks passed!`
4. `/opt/homebrew/bin/python3.11 -m pytest -q`
   - exit `0`
   - `507 passed, 18 skipped in 94.14s`
5. `/opt/homebrew/bin/python3.11 -m build`
   - exit `0`
   - built `cocoaskills-0.12.6.dev0+g6fc2fd97d.d20260729.tar.gz`
   - built `cocoaskills-0.12.6.dev0+g6fc2fd97d.d20260729-py3-none-any.whl`
6. `git diff --check`
   - exit `0`
   - no output

The task-local mypy environment was created reproducibly with:

```text
uv venv /Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/mypy-venv --python /opt/homebrew/bin/python3.11
uv pip install --python /Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/mypy-venv/bin/python mypy -e .
```

## Findings and scope notes

- Caller-defined class keys keep transaction infrastructure generic; ordered numeric prefixes encode policy such as consumer-last without coupling this slice to unfinished installer/compiler parsing.
- Lock ordering is enforced per execution thread and process-wide at the home-lock boundary, while independent projects may still hold already-acquired project locks during serialized home mutation.
- Corrupted journal paths are treated as implementation corruption and preserved for diagnosis; recovery does not follow redirected cleanup paths or overwrite unknown live bytes.
- The requested `logbook` executable is not installed (`logbook --help` exit `127`), the board CLI exposes no logbook command, and tool discovery found no logbook connector. These findings are therefore persisted in this outcome and task notes.
- No staging, commit, publication, pin change, Go UX, compiler policy, or installer integration was performed.
