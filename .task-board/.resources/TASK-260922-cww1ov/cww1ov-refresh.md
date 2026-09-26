# TASK-260922-cww1ov base refresh → revision 6 (orchestrator brief, binding)

Revision 5 was ACCEPTED, but its integrate refused `integration_base_moved`: trunk advanced to `fad88136`
(2qvzwk + 1bfk8y landed) with a change to `.github/ci/platform-cases.tsv`, which revision 5 also changes. The acceptance was
released (`worktree invalidate-acceptance`). NOTHING about the content is in question — base refresh only.

1. `task-board worktree refresh-candidate TASK-260922-cww1ov`. `CHANGELOG.md` is `merge=union` in `.git/info/attributes`.
   Any other conflicting path (e.g. `.github/ci/platform-cases.tsv`): resolve by keeping BOTH sides' rows
   (trunk's new rows + yours), in the file's own ordering convention, then follow the command's own
   `--replay-resolutions` instructions — never hand-commit the replay worktree.
2. Prove the refreshed candidate equals revision 5 except for the merged paths: `git status --short` and a
   per-file diff summary in results.
3. Re-run your narrow rows on the refreshed tree (bounded calls, real exit codes); for platform-cases.tsv also
   `sh .github/ci/ledger-consistency.sh` and `sh .github/ci/gate-selftest.sh`.
4. Append "Revision 6 (refresh)" to results, then `task-board handoff TASK-260922-cww1ov --role developer`. A
   `run_wrote_outside_worktree … policy warn` block is a warning — verify the status moved to `to-review`.
No product change, no test weakened. No LOGBOOK.md.
