# Reviewer brief

Review `BUG-260810-2oxt8b` independently. Do not edit product code, commit, push, open, or merge pull requests. Do not edit `.task-board` files directly; every board change must use `task-board` commands.

Verify the task-scoped reconciliation outcome, the evidence resources on the historical items, the cited GitHub states/checks, and the patch-equivalence inference for PR #1.

If the evidence is accepted, apply these terminal board transitions as the reviewer verdict work:

- `TASK-260713-c7a18d` (`harden-shell-activation`) -> `done`;
- `STORY-260713-b4e219` (`seamless-manager-lifecycle`) -> `done`;
- `TASK-260713-7a9c1e` (`review-and-correct`) -> `done`;
- `STORY-260713-72b914` (`production-profile-conformance`) -> `done`.

Confirm the expected parent aggregation. Add the PR #5 evidence resource to `STORY-260713-72b914` if the CLI permits it.

The producer found that `STORY-260713-72b914/progress.md` stores the legacy raw status `in-progress`: structured reads project `backlog`, but current mutations fail while parsing the raw value. First look for a supported task-board normalization path. Never repair it by direct file edit. If the installed CLI genuinely cannot normalize it, request changes by returning `BUG-260810-2oxt8b` to `analysis` with a task-scoped reviewer verdict that identifies the exact missing CLI behavior and the required rework; this is recoverable and is not a stop-the-line blocker.

Accept only if:

- all four historical items reach `done` through CLI mutations;
- evidence remains attached and correct;
- `EPIC-260713-c12fbe` stays `backlog` with no children;
- `EPIC-260810-271m92` and `STORY-260810-rn6fg1` stay `backlog`;
- `task-board validate --json` reports no errors or warnings.

Persist a task-scoped reviewer verdict resource before choosing exactly one reviewer verdict branch.
