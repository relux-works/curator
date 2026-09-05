# Producer brief: publish Change Request revision 2 for the schemas/vectors batch (bring the rebased head into the fresh workspace)

The reviewed and rebased work is `fd237ba` on `draft/environments-schemas-1-1` (one signed commit past curator-spec main
`f61ee9a`; PR #42 all checks green). The previous managed Story workspace (base `a68559b`) was discarded because main
moved. Your run has a fresh workspace at trunk. Do exactly this:

1. In your managed story worktree: `git status --short` empty; `git log --oneline -1` shows `f61ee9a` (or later main).
   Run `git cherry-pick -S fd237ba` (`git cat-file -t fd237ba` = commit). If it conflicts, stop and report — do not
   resolve by hand.
2. Verify: `git log --oneline -2` = your new commit then the base; `git log --show-signature -1` verifies;
   `git diff fd237ba HEAD --stat` empty; `git status --short` empty (remove `tools/__pycache__/` if present; do not run
   the validator).
3. Do NOT edit any file otherwise. `task-board handoff TASK-260905-1xkxe4 --role developer` publishes CR revision 2.
   Attach `TASK-260905-1xkxe4_cr2-publish-report.md` with the handoff output (revision, candidate commit and tree) or
   the exact refusal. Never write LOGBOOK.md or anything into the control root or the repository.
