# TASK-260720-z2z795 final-review rework outcome

Date: 2026-07-30
Role: developer
Run: `RUN-260729-e44939`

## Provenance and scope

- Accepted dependency: `TASK-260720-1pvfj5` (`done`).
- Original task base: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Rework starts from the exact independently reviewed task commit:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`.
- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`.
- The rework remains unstaged and uncommitted on
  `task/TASK-260720-z2z795-transaction-engine`.
- Exactly four task-owned paths are modified:
  `src/csk/locking.py`, `src/csk/transactions.py`,
  `tests/test_locking.py`, and `tests/test_transactions.py`.
- No SSH host, Tailscale route, host application, other repository, pin,
  installer policy, Go UX, stage, commit, or publication was touched.

## Review finding 1: process-wide lock ordering

`locking.py` now keeps pending and held project/build/home ownership in a
process-wide registry keyed by the canonical physical manager-home identity.
Reservations are installed before filesystem acquisition and are removed on
every success and failure path.

The hierarchy now guarantees:

- a held or pending build lock blocks manager-home acquisition across threads;
- a pending project acquisition blocks manager-home acquisition;
- a pending or held manager-home lock blocks new project/build acquisition;
- concurrent manager-home consumers still serialize through the filesystem
  lock instead of failing, preserving both-project success behavior;
- textual/symlink aliases of the same manager home share one process state.

Deterministic regressions exercise held and pending cross-thread transitions,
canonical home aliases, failure cleanup, and concurrent home consumers.

## Review finding 2: namespace independence

Planning and journal validation now compare every live, desired-staging,
backup, rollback, and cleanup-tomb namespace against every other target and
the complete journal tree before journal or staging mutation.

Comparison uses physical resolution, existing-object identity, component-wise
ancestor checks, Windows case behavior, and macOS case/canonical-Unicode
behavior. It fails closed when physical identity cannot be validated.

Regressions cover:

- journal ancestor and descendant targets;
- target parent/child overlap;
- live paths equal to another target sidecar;
- symlinked parent/child aliases;
- platform case aliases;
- macOS canonical-Unicode aliases.

Every rejection asserts that no journal, sidecar, or live-target residue is
created.

## Review finding 3: durable, journal-owned cleanup

The durable journal now records cleanup progress for every staging, backup,
and rollback sidecar:

- target index and sidecar role;
- canonical sidecar and deterministic cleanup-tomb paths;
- transaction-owned full digest;
- deterministic per-file/per-directory manifest;
- cleanup state (`pending`, `tombed`, or `removed`).

Before deletion, the canonical sidecar is atomically moved to its owned cleanup
tomb. Windows routes this move through `MoveFileExW` with
`MOVEFILE_WRITE_THROUGH`. The recoverable journal remains canonical while each
entry is removed and progress is durably replaced; only fully removed cleanup
state can transition to the final journal tomb.

Recovery:

- completes a crash after sidecar tomb publication for commit and rollback;
- resumes deterministic partial directory removal;
- replays a crash immediately before and after final journal tomb publication;
- rejects changed or unrecorded cleanup bytes without deleting them;
- discards a safely journaled partial staging copy interrupted during prepare.

## Final local validation

Every command below exited `0`:

- focused pytest: `57 passed, 1 skipped`;
- strict mypy: no issues in `57 source files`;
- changed-file Ruff lint: clean;
- changed-file Ruff format: all four files formatted;
- full pytest: `627 passed, 20 skipped`;
- package build: sdist and wheel built;
- `git diff --check`: clean.

The exact commands, outputs, real exit codes, expected-red development runs,
final hashes, and worktree status are in
`TASK-260720-z2z795_final-review-rework-validation.log`.

## Explicit platform boundary

The final focused run skips one real-Windows file/tree flush test on macOS.
Windows `MoveFileExW` flag routing and durable cleanup state transitions are
covered deterministically locally. A real Windows run was not executed on
these unstaged bytes because the active orchestrator directive explicitly
forbids SSH/host access and the task instructions forbid staging, committing,
or publishing. The prior green GitHub Windows matrix covers
`5f8bfbbd...`, not these rework bytes, and is not claimed as current-byte
evidence. Reviewer/orchestrator owns any required windows-latest publication
gate after this developer handoff.

The standalone `logbook` command remains unavailable in this environment.
Findings and decisions are persisted in this outcome, the validation ledger,
and task notes.
