# Producer brief: publish the Change Request for the manager/cli batch (no edits)

The work is committed on the story branch `task-board/story/STORY-260905-2z9pw4` at `9af8af8`: ONE signed
commit past base `a68559b` (the earlier two-commit series plus a stray `LOGBOOK.md` were squashed by the
orchestrator into one commit with the tree of `ffbf803`; `LOGBOOK.md` is not part of it). The draft branch
`draft/environments-manager-cli-1-1` mirrors the same commit.

Do exactly this: in your managed story worktree run `git status --short` (must be empty — if
`tools/__pycache__/` reappears, `rm -rf` it; do not run the validator), `git log --oneline -2` (must show
`9af8af8` then `a68559b`), `git log --show-signature -1` (must verify). Do NOT edit, commit, reset, or rebase.
Then `task-board handoff TASK-260905-369vye --role developer` — this publishes the Change Request. Attach
`TASK-260905-369vye_cr-publish-report.md` with the handoff output (revision, candidate commit and tree) or the
exact refusal. Never write LOGBOOK.md or anything else into the control root or the repository.
