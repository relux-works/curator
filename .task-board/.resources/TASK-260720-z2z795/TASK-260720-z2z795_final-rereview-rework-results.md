# TASK-260720-z2z795 final re-review rework

## Outcome

The focused rework closes the three findings in
`TASK-260720-z2z795_final-rereview-verdict.md` without expanding beyond the
four task-owned Python paths.

## Lock reservation and handoff changes

- Failed same-instance acquisition now leaves the pre-existing held
  reservation untouched. Failure cleanup applies only after the current
  attempt has successfully created a pending reservation.
- Project, build, and home pending/held states remain separate on every tested
  success and failure path.
- Manager-home ownership is process-wide per canonical home and set-valued.
  A consumer that acquires the filesystem lock during the prior consumer's
  unlink/process-release window can enter without a spurious order error.
- Project and build acquisitions remain blocked throughout all pending, held,
  and serialized handoff states.
- Deterministic re-entry, release-window, and delayed-release concurrent stress
  regressions exercise these boundaries.

## Durable staging changes

- Preparing journals now record the canonical desired source, a deterministic
  per-entry file/directory manifest, current entry index, entry ownership, and
  confirmed plus write-ahead byte/prefix-digest progress.
- Entry intent is durably saved before sidecar creation. Each file chunk is
  write-ahead journaled before writing and fsync, then confirmed in the journal
  after durable write.
- Recovery validates completed entries exactly and active partial files against
  the durable source manifest, confirmed prefix, or one journaled write-ahead
  range before it records cleanup ownership.
- Changed, replaced, or newly added post-crash file/directory bytes are
  rejected and preserved with the journal. Valid partial staging is cleaned,
  and a fully completed manifest can be cleaned even if its private source has
  disappeared.
- Existing namespace-isolation and journal-owned Windows write-through sidecar
  cleanup behavior remains covered by the focused and full suites.

## Validation

- Reviewer-regression pytest: 14 passed, exit 0.
- Original contract-targeted pytest: 13 passed, exit 0.
- Focused locking/transaction pytest: 69 passed, 1 Windows-only skipped,
  exit 0.
- Strict mypy: 57 source files, exit 0.
- Ruff lint and format checks: exit 0.
- Full pytest: 639 passed, 20 skipped, exit 0.
- Package build: sdist and wheel built, exit 0.
- `git diff --check`: exit 0.

Exact commands, outputs, hashes, provenance, and platform boundaries are in
`TASK-260720-z2z795_final-rereview-rework-validation.log`.

## Boundary

The worktree remains unstaged and uncommitted on
`5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`. No SSH, commit, push, `wb`,
tag, release, or current-byte real-Windows run was performed, as required by
the run directive.
