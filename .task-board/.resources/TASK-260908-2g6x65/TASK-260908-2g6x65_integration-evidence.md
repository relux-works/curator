# Integration evidence — TASK-260908-2g6x65 / CR-TASK-260908-2g6x65-1 rev1

Run: RUN-260908-df687d (developer/implementer integration owner). Date 2026-09-08.
Control root: /Users/iv/Developer/ReluxWorks/.worktrees/launcher-control, inherited TASK_BOARD_CONFIG.

## Commands and real exit codes

| Step | Command | Exit |
|---|---|---|
| 1 | `git -C /Users/iv/Developer/ReluxWorks/curator -c pull.rebase=false pull --ff-only origin main` | 0 ("Already up to date", HEAD 4f980b1a) |
| 2 | `task-board --no-update-check --board-dir /Users/iv/Developer/ReluxWorks/curator/.task-board worktree complete STORY-260908-xadoax --cr TASK-260908-2g6x65 --revision 1 --landed-commit 25379f22245e3bd8d2b81b192287d04368bf42c7 --commit-time 2026-09-07T21:30:00+03:00 --json` | 0 |

Raw stdout/stderr: `.temp/TASK-260908-2g6x65/complete-01.{stdout,stderr}` (stderr empty).

## Result of step 2

- txn_id `STORY-260908-xadoax/CR-TASK-260908-2g6x65-1/1`, phase `cleanup_pending`
- story_commit_oid `25379f22245e3bd8d2b81b192287d04368bf42c7`
- board_commit_oid `22c0ce85b4253ca7b1ed3ec818123fc962cc3da1`, board_published true, ref `refs/heads/main`
- revalidated: true (tool ran configured `make check` against the accepted tree; supported real revalidation, no CI)
- manifest: 17 board paths written (15 lane, 2 shared), digests in stdout log

## Verification

- Landed commit `25379f22` is a signed (`G`, Ivan Oparin) commit, ancestor of launcher `origin/main` (== `25379f22`); its tree is `ae4818100f2649cd833629bde07950f8f5e3843e`, identical to CR `candidate_tree_oid`.
- Board commit `22c0ce85` "Record STORY-260908-xadoax board state": `git verify-commit` good ECDSA signature for oparin@me.com (key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM); curator HEAD == origin/main == `22c0ce85`; contained in origin/main.
- `get(STORY-260908-xadoax){status}` → done; `get(TASK-260908-2g6x65){status}` → done.
- CR record unchanged: state accepted, kind story_final, base `484933b3`, producer RUN-260907-9b71c2, reviewer RUN-260907-17f39f.
- No spawn directives for RUN-260908-df687d.

## Not touched

No code/LOGBOOK/control-root edits, no new CR, no signer/trust/CR/tree-proof edits, no installs, daemon restarts, tags, releases, ax operations or runtime-home edits. Pre-existing unrelated dirty board activity files in the curator checkout preserved. `worktree cleanup` is now eligible per tool note; left to the parent.
