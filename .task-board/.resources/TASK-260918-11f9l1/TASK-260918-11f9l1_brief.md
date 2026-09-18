# Brief — TASK-260918-11f9l1: replay the accepted rc.12 union onto the moved trunk (story-final)

Story `STORY-260917-3w3lvj` (conformance-pin-rc12-promotion), `EPIC-260910-2hw1xb`.
State of the story: `TASK-260917-16l2md` (the union: `SPEC_PIN` → `dced9b8`
+ E2/E4/S4 manager work) is accepted at revision 5 and checkpointed as
commit `73fc8a4` on the Story branch (base `3c45d4b`); `TASK-260917-2ecpjv`
(rc.12 qualification, verdict qualified) is accepted or in review. This
task is the LAST open leaf: your handoff constructs the story-final Change
Request against the fresh trunk authority, so the checkpoint must first be
replayed onto it. Rules: `remediation-manager-producer-rules.md` (attached
to the epic; rule 8: never build into the worktree). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260917-3w3lvj/worktree`
(branch `task-board/story/STORY-260917-3w3lvj`, currently at `73fc8a4`).

## Why the trunk conflicts
Since `3c45d4b`, curator `main` gained the S6 story squash `64cacfc`
(shell-hook trust rows: `cmd/curator/envstatus.go`, `cmd/curator/main.go`,
`internal/envprofile/status.go`, `internal/shell/*`, `CHANGELOG.md`),
`d00fe7a` (rose-air rustup lane: `.github/ci/gate-selftest.sh`,
`.github/ci/install-rust-toolchain.sh`) and board-state commits. A dry
rebase of `73fc8a4` onto `main` conflicts in exactly three files, all
additive (both sides added members/rows at the same place):
- `internal/envprofile/status.go` — the `Status` struct: S6 adds
  `ShellHookTrust` / `ShellHookTrustWarnings`; the union adds `Providers`,
  `S4Profile`, `PassableEnvNames`, `Warnings`, `MCPDeclarations`. Keep ALL,
  gofmt-aligned; state the field order.
- `cmd/curator/envstatus.go` — the `env status` printer: S6 prints the
  shell-hook warnings and rows; the union prints the S4 profile line, the
  machine-level warnings and the per-scope MCP declaration rows. Keep both
  blocks; choose and STATE the closed output order (recommended: the
  union's §12 posture block first — s4_profile, warnings, mcp-declarations,
  provider rows — then the S6 shell-hook block, matching the order of the
  manager §10 inventory in the landed spec: hook-trust is the first gate
  there, so if the spec's order is the rule, put the S6 block FIRST; read
  `profiles/manager.md` §10 at the rc.12 root and follow it), and make the
  tests of both sides agree with that order.
- `CHANGELOG.md` — two Unreleased entries; keep both (S6 entry as on
  `main`, the union's entries as in the checkpoint), no rewording.
`cmd/curator/main.go` auto-merges (S6's `curator status` warning line and
the union's changes are in different hunks) — still read the merged
result: `curator status` must carry both the shell-hook rows and the
union's posture rows in the same order as `env status`.

## Do exactly this
1. `task-board worktree obligations` — confirm no obligation row for this
   story other than review/checkpoint bookkeeping; `git -C <worktree> status --short`
   must be clean at `73fc8a4`.
2. `task-board worktree refresh-candidate TASK-260918-11f9l1` — the first
   attempt stops at the three regular-file conflicts and retains a replay
   worktree with `resolution-template.json` in its parent directory. Prepare
   the replacement content for each conflicted path (the union of both sides
   as described above, gofmt-clean), fill `--replay-resolutions` exactly as
   the template asks (bound to `REBASE_HEAD`, the unmerged path and the
   SHA-256 of the replacement bytes), and re-run
   `task-board worktree refresh-candidate TASK-260918-11f9l1 --replay-resolutions <file>`.
   Never hand-commit the retained replay worktree; never edit workspace
   registry records; never `git rebase` the Story branch yourself. Quote
   every command and output in the results.
3. After the replay: in the Story worktree, run the union's and S6's tests
   at the rc.12 root
   (`export CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1`, create
   the detached root with `git -C …/curator-spec worktree add /tmp/spec-rc12 dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
   if absent): `go build ./... && go vet ./... && gofmt -l .` and the full
   `make ci-test` (or the documented equivalent). Fix ONLY what the
   combination broke (golden output order in `cmd/curator` tests,
   struct-order assertions, JSON field expectations), as an uncommitted
   working-tree delta on top of the replayed checkpoint — that delta is
   this task's Change Request. Do not touch the pin, the union's behaviour
   or S6's behaviour.
4. Identity proof in `TASK-260918-11f9l1_results.md`: per-file
   `git patch-id --stable` of `replayed checkpoint vs main` against
   `73fc8a4 vs 3c45d4b` — every file identical except the three resolved
   files (quote their merged regions); list your fix-up delta file by
   file; the closed output order for both commands; gate transcripts.
   Confirm rule 8 (`git status --short --untracked-files=all` shows no
   build output).
5. Tick the checklist, then `task-board handoff TASK-260918-11f9l1 --role developer`
   immediately after the replay + fixes (the story-final Change Request is
   constructed against the trunk at that moment; if it refuses with
   `stale-anchor` because the trunk moved again during your run, quote the
   error, re-run step 2 (refresh-candidate replays from the unchanged
   original checkpoints) and hand off again — at most twice, then stop and
   report).
