# Refresh-and-handoff run — TASK-260910-3ungjy (S6 approval commands, revision 3)

The revision-3 rework is complete in the Story worktree
`<control-root>/.temp/STORY-260910-2awkzu/worktree` (uncommitted, 12 files;
`TASK-260910-3ungjy_results.md` already carries the "Revision 3" section
written by the previous run). Its handoff was refused twice with
`stale-anchor: change_request_base_authority_mismatch`: the workspace
checkpoint `1e6a165` (the replayed `TASK-260910-1952mz` checkpoint) sits on
trunk `aa46ecd`, while trunk has since advanced to `0f0ae61` through
board-state commits that touch no code path. The orchestrator will not move
trunk again until this handoff is published.

Do exactly this:
1. `git -C <worktree> status --short` and `git -C <worktree> log --oneline -2`
   — confirm the 12 uncommitted rev-3 paths and tip `1e6a165`; do NOT stage,
   stash, commit or reset anything.
2. `task-board worktree refresh-candidate TASK-260910-3ungjy` — replays the
   checkpoint onto fresh trunk with signing in an isolated workspace and
   preserves the working candidate exactly. Quote the output. If it reports a
   regular-file conflict, prepare `--replay-resolutions` from the retained
   replay worktree's `resolution-template.json` ONLY for conflicts inside
   files the rev-3 work touches, choosing the union that keeps both trunk's
   and the checkpoint's hunks; otherwise quote the refusal and stop.
3. After the refresh: `git -C <worktree> log --oneline -3` (the checkpoint now
   descends from `0f0ae61`), `git -C <worktree> status --short` (same 12
   paths), then re-run the narrow gates with
   `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`:
   `go build ./... && go vet ./... && gofmt -l internal cmd && go test -count=1 ./cmd/curator/... ./internal/hookapproval/... ./internal/envprofile/... ./internal/shell/...`
   (`set -o pipefail`, exit codes). Append a short "Revision 3 — refreshed
   base" note with the transcripts to `TASK-260910-3ungjy_results.md`
   (update the resource).
4. Tick the checklist items the rev-3 work satisfies (verify each against
   the code, do not tick blindly), then `task-board handoff TASK-260910-3ungjy --role developer`.
   The runtime runs the hosted gate. No other changes.
