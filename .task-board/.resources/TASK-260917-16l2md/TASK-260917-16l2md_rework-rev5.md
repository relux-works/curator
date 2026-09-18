# Rework brief — TASK-260917-16l2md, revision 5 (rc.12 pin promotion union)

Revision 4 closed all three round-1 corrections; the reviewer found ONE
remaining defect (`TASK-260917-16l2md_review-verdict-rev4.md`): the candidate
accidentally captured a local build output — the 18.8 MB Mach-O executable
`curator` at the repository root (untracked file in the Story worktree). It
is the only extra path between revisions 3 and 4; everything else is
accepted.

Do exactly this:
1. `git -C <worktree> status --short` — confirm the untracked root
   `curator` binary; delete it (`rm <worktree>/curator`); confirm no other
   build output or stray file is present (`git status --short` shows only
   the source/test/docs/CI changes of the union).
2. Append a "Revision 5" note to `TASK-260917-16l2md_results.md`
   acknowledging the cleanup (and correcting §11.6's file count); update the
   resource.
3. `task-board handoff TASK-260917-16l2md --role developer`. No other
   changes; never build into the worktree root again — use `/tmp` or the
   ignored `bin/` directory.
