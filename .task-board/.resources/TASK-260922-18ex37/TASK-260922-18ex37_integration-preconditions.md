# TASK-260922-18ex37 integration preconditions (bound run, rev4)

Binding: CR-TASK-260922-18ex37 revision 4 ACCEPTED (story_final) per orchestrator instruction `18ex37-integrate-land.md`. This run did NOT execute `task-board worktree integrate` or `worktree checkpoint` and changed no file, per the Integration Assignment footer: the runner performs the bound landing synchronously after this run. No `set_status` and no generic `handoff` were issued; board left at `integrating` for the landing transaction (the only writer allowed to set `done`). No `.temp/integrate-18ex37-land.log` was produced here by construction.

Preconditions confirmed (all commands exit 0, standalone, no pipes hiding status):
- TASK-260922-18ex37 status=integrating (task-board q get, exit 0).
- STORY-260922-2goxjs status=integrating (task-board q get, exit 0).
- Worktree branch: task-board/story/STORY-260922-2goxjs (git rev-parse, exit 0).
- HEAD: 12a6e5bf (TASK-260922-3bbvrs landed; the 3bbvrs checkpoint the Story carries) — no commit of 18ex37 by this run, so no `change_request_candidate_committed_past_checkpoint` risk introduced here.
- Worktree dirty: 59 porcelain paths = uncommitted CR-4 candidate (git status --porcelain=v1 | wc -l, exit 0 via pipefail shell).
- SPEC_PIN in .github/workflows/ci.yml: dcc7f015e2d97edf2d52928afb6fd79ec8129e8b (grep, exit 0) — matches operator decision keep-dcc7f015 / #89/v9 OUT.
- Spawn directives for RUN-260925-c9663b: none (exit 0).

Not independently re-verified in this bound run: hosted-gate validation log contents and accept_cr ledger entry — taken from the orchestrator ACCEPTED statement; the synchronous landing transaction will refuse if stale, and that refusal (if any) belongs to the orchestrator delivery step.
