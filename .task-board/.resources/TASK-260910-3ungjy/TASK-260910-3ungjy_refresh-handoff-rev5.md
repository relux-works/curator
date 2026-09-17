# Refresh-and-handoff run — TASK-260910-3ungjy (S6 approval commands), revision 5

Revision 4 was ACCEPTED by review (`TASK-260910-3ungjy_review-verdict-rev4.md`)
but its integration was refused with `integration_base_moved`: trunk advanced
with `STORY-260910-3vxe3y` (`6645b9b`, package-lock-and-frozen-resolution)
which changed `cmd/curator/main.go`, a path revision 4 also changes; the
board demoted the revision to `stale` — "no one has looked at the
combination". The workspace checkpoint (`dc5675e`, the replayed
`TASK-260910-1952mz` leaf) is not contained in the current authority
(`b92bf5e`), so `worktree converge` refuses and the remediation is the
producer-run `refresh-candidate`. The rev-4 work is intact and uncommitted
in the Story worktree `<control-root>/.temp/STORY-260910-2awkzu/worktree`.

Do exactly this:
1. `git -C <worktree> status --short` and `git -C <worktree> log --oneline -2`
   — confirm the uncommitted rev-4 paths and tip `dc5675e`; do NOT stage,
   stash, commit or reset anything.
2. `task-board worktree refresh-candidate TASK-260910-3ungjy` — replays the
   checkpoint onto fresh trunk with signing in an isolated workspace and
   preserves the working candidate exactly. Quote the output. If it reports
   a regular-file conflict (expected candidate: `cmd/curator/main.go`, where
   trunk's package-lock work and this task's `hook` subcommand / status rows
   meet), prepare `--replay-resolutions` from the retained replay worktree's
   `resolution-template.json`: the resolution is the UNION that keeps both
   trunk's hunks and the checkpoint's hunks; bind each resolution to
   REBASE_HEAD, the unmerged path and the SHA-256 of the replacement content
   as the template requires; re-run with `--replay-resolutions`. Never
   hand-commit the retained replay.
3. After the refresh: `git -C <worktree> log --oneline -3` (checkpoint now
   descends from the fresh authority), `git -C <worktree> status --short`
   (same rev-4 paths, plus any merge of the working delta against the new
   `cmd/curator/main.go` — if the working-tree delta itself conflicts, resolve
   it in the worktree as the union and say exactly what you changed); then
   the narrow gates with
   `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`:
   `go build ./... && go vet ./... && gofmt -l internal cmd && go test -count=1 ./cmd/curator/... ./internal/hookapproval/... ./internal/envprofile/... ./internal/shell/...`
   (`set -o pipefail`, exit codes).
4. Append a "Revision 5 — refreshed onto 6645b9b/b92bf5e" section to
   `TASK-260910-3ungjy_results.md` (update the resource) naming the exact
   combination points in `cmd/curator/main.go` (file:line of trunk's hunks
   next to yours) so the reviewer can look at the combination; tick the
   checklist items the work satisfies (verify, do not tick blindly); then
   `task-board handoff TASK-260910-3ungjy --role developer`. The runtime
   runs the hosted gate. No other changes.
