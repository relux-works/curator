# Brief — TASK-260918-fjl6v2: land the accepted rc.12 union on the current trunk (story-final)

Story `STORY-260918-2yvd86` (conformance-pin-rc12-landing), `EPIC-260910-2hw1xb`.
Why this task exists: the union `TASK-260917-16l2md` (SPEC_PIN → `dced9b8`
= tag `v1.0.0-rc.12`, qualified by `TASK-260917-2ecpjv`; E2/E4/S4 manager
work) was accepted at revision 5 and checkpointed as `73fc8a4` on the old
Story branch (base `3c45d4b`); the curator trunk then moved (S6 squash
`64cacfc`, rustup CI fixes, board-state commits), and the sanctioned replay
(`worktree refresh-candidate`) fails with a tool error on that story. Your
predecessor `TASK-260918-11f9l1` worked out and validated the combination in
a scratch tree and attached everything you need (its resources). This
story's workspace is a FRESH fork of the current trunk, so the story-final
Change Request can be constructed. Rules: `remediation-manager-producer-rules.md`
(attached; rule 8: never build into the worktree). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260918-2yvd86/worktree`
(branch `task-board/story/STORY-260918-2yvd86`, provisioned at spawn from
the current `main`; record its base commit in the results).

## Inputs (all attached to this task)
- `rc12-union-73fc8a4-vs-3c45d4b.patch` — the accepted union delta
  (`git diff 3c45d4b 73fc8a4`, 40 files, patch-id `ccc9574a`).
- `TASK-260918-11f9l1_resolved-CHANGELOG.md`, `…_resolved-envstatus.go`,
  `…_resolved-status.go` — the three conflict resolutions (union of both
  sides; S6 shell-hook block FIRST, then the union's §12 posture block, per
  the manager §10 inventory order; see `TASK-260918-11f9l1_results.md` §2–§3).
- `TASK-260918-11f9l1_fixup.patch` — test-only fix-up (+33 lines in
  `cmd/curator/hook_posture_test.go` and `cmd/curator/hook_test.go`: the S6
  tests plant stub `curator-run`/`curator-session` providers because the
  union's §12 provider rows make a missing provider non-current).
- `TASK-260918-11f9l1_results.md` — identity proof, output order, gate
  transcripts (green at trunks `6401d3c` and `1c464c5`).

## Do exactly this
1. `git -C <worktree> status --short` clean; `git log --oneline -1` = the
   current trunk. `git apply --3way --index rc12-union-73fc8a4-vs-3c45d4b.patch`
   — expect conflicts ONLY in `CHANGELOG.md`, `cmd/curator/envstatus.go`,
   `internal/envprofile/status.go` (if the trunk moved further and a new
   conflict appears, resolve it as the union of both sides and document it
   the same way). Replace the three conflicted files with the resolved
   versions byte-for-byte (verify SHA-256 against the results §6 table),
   `git add` them, then `git apply --index TASK-260918-11f9l1_fixup.patch`.
   `git reset -q` afterwards so the delta is an uncommitted working-tree
   change (new files intent-to-added with `git add -N`); never commit on
   the Story branch.
2. Identity proof (results): per-file `git patch-id --stable` of
   `git diff HEAD -- <file>` vs the union patch for each of the 40 files —
   identical except the three resolved files and `cmd/curator/main.go`
   (auto-merge: show it equals checkpoint + S6 hunks); plus the fix-up
   files. State the closed output order of `curator status` / `env status`.
3. Gates at the rc.12 root (`export CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1`,
   create with `git -C …/curator-spec worktree add /tmp/spec-rc12 dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
   if absent): `go build ./... && go vet ./... && gofmt -l .` and the full
   `make ci-test` (or the documented equivalent) — quote exit codes; builds
   go to `/tmp` (rule 8).
4. Attach `TASK-260918-fjl6v2_results.md` (write it OUTSIDE the worktree),
   tick the checklist, confirm `git status --short --untracked-files=all`
   lists only the union's 40 paths + the 2 fix-up test files, then
   `task-board handoff TASK-260918-fjl6v2 --role developer` IMMEDIATELY (the
   story-final Change Request is built against the trunk at that moment).
   If it refuses with `stale-anchor`, quote it, run
   `task-board worktree refresh-candidate TASK-260918-fjl6v2` (this story
   has a single, fresh checkpoint — the replay should work; resolve any
   conflict as a union via `--replay-resolutions`) and hand off again; at
   most twice, then stop and report.
