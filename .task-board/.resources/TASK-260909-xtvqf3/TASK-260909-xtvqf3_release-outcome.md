# TASK-260909-xtvqf3 public release outcome — stranded accepted task_delta invalidated after PR199 repair

Run: RUN-260909-cdae9b (role developer, archetype implementer, muse-spark xhigh) — bound recovery run, 2026-09-09.
Scope: release-only. No code, gate, install, daemon, CI, tag, ax, LOGBOOK, or board-state-push actions.

## 1. Pre-recovery evidence (all read-only, exit 0 unless noted)

- Installed binary: `task-board version 0.24.3-335-gac9044a5 (commit ac9044a5, built 2026-09-09T16:35:07Z)` at /Users/iv/.local/bin/task-board. Matches reviewed PR199 source per launcher .temp/resume-2026-09-09/pr199-delivery-install.md (PR merged 2026-09-09T16:34:36Z, accepted tree b09b9f51d896f7154166f30d44a33a7235299480, setup.sh exit 0, daemon PIDs unchanged).
- Bound state reobserved: TASK-260909-xtvqf3 `integrating`, STORY-260909-1cw05m `integrating`, sibling TASK-260909-3d1589 `integrating` with CR rev 3 `checkpointed` (empty delta). Story worktree clean (`git status --porcelain` empty), HEAD 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3, lease held by RUN-260909-cdae9b (this run). `worktree status` exit 0.
- Prior refusal read: TASK-260909-xtvqf3_integration-refusal.md (RUN-260909-16fac6) — `worktree complete --revision 3` refused exit 1 with `integration_blocked: task_delta, not story_final`.
- Public help read: `worktree invalidate-acceptance --help` documents this exact stranded-kind exit (accepted record whose re-derived kind differs). Note: no `references/change-request-lifecycle.md` file exists under the frozen launcher root or source checkout (bounded filename search, exit 0, no match); the installed public help text is the authority relied upon.
- Pre-recovery SHA256 (frozen control root /Users/iv/Developer/ReluxWorks/.worktrees/launcher-control/.temp/changerequests/TASK-260909-xtvqf3/):
  - rev-000003.json: 693e8797d0383a4b50e42c6b34c8783bdd27e08fee79abf3345761786bc8ddcc
  - current.json:    1e073642b852f1acd9632acc462e5f5a340b6df05e5a57e5436ae22bbc7ae1c7
  - events.jsonl:    889f55e0038372eccae7457235f73fe02360349c67495aa8f7b8fee9498be7b8
  - rev-000001.json: fe3af263e187d840426bf5105a10d5ae992357d1519badd08238be8e5e8d5c5d
  - rev-000002.json: 28a435125cc264ebb3595080055fa81d6580b3f3bd74c6f1aabab4d34bca7179

## 2. Public release command and exact receipt (exit 0)

Command (from story worktree, branch task-board/story/STORY-260909-1cw05m):

    task-board --no-update-check worktree invalidate-acceptance STORY-260909-1cw05m --cr TASK-260909-xtvqf3 --reason "accepted task_delta no longer matches derived story_final after reviewed bound checkpoint loader repair" --json

Exit: 0. Exact stdout JSON:

    {"change_request_state": "accepted", "element_id": "TASK-260909-xtvqf3", "revision": 3, "status": "to-dev"}

No forced transition was needed; the command succeeded on first attempt. No manual set_status, no CR manufacture, no private-record edit.

## 3. Post-release proof: accepted history byte-identical

Re-hashed after release (exit 0) — all five digests identical to §1:

- rev-000003.json: 693e8797…cc (unchanged, 1526 bytes, mtime still 2026-09-09 16:06:01)
- current.json: 1e073642…c7 (unchanged, still revision 3)
- events.jsonl: 889f55e0…b8 (unchanged, still ends at seq 4 accepted by RUN-260909-2fc911)
- rev-000001/2.json unchanged.

Reobserved board state (exit 0): TASK-260909-xtvqf3 `to-dev`; STORY-260909-1cw05m `integrating`; sibling TASK-260909-3d1589 `integrating`/checkpointed. Worktree still clean at 3ff66a9; lease still held by RUN-260909-cdae9b.

## 4. Remaining steps (parent-owned, outside this run)

1. Parent spawns a NEW ordinary developer run for TASK-260909-xtvqf3 (now `to-dev`, rework reachable) to publish the next story_final revision. Runtime publication validation belongs to that producer.
2. Independent Astra medium acceptance of the new revision (reviewer run, accept_cr).
3. Signed bound Complete via `worktree complete STORY-260909-1cw05m --cr TASK-260909-xtvqf3 --revision <N>` (still without --landed-commit; artifact-only scope), including curator board-owner `git pull --ff-only` convergence and signed board-state publication. Note: story lease is currently held by terminal RUN-260909-cdae9b; release via `worktree gc` when the holding run is provably terminal.
4. Original diagnostics candidate, all accepted gate artifacts, and CR history preserved byte-for-byte throughout; nothing above redevelops gates or touches launcher source.
