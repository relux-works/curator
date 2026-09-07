# A0 completion attempt 01: `worktree complete` refused (board repository has two remotes)

Run: RUN-260908-70092d, role researcher (integration owner for accepted CR-TASK-260908-qblycn-1 rev 1, repository_delta=empty).
task-board: `0.24.3-317-g416693dd` (commit 416693dd, PR 187 support present: `worktree complete` exists and its help documents the empty-delta path with no `--landed-commit`).

## Pre-flight
- `git -C /Users/iv/Developer/ReluxWorks/curator -c pull.rebase=false pull --ff-only origin main` -> "Already up to date." Curator main HEAD = `d1bb0a4e92304526c782f77607aa30abc2f3ae07` (origin/main "Keep a release candidate out of the install channels"; this is past the 04550e2 named in the brief). Unrelated dirty board/LOGBOOK state untouched (git status shows ~20 modified/deleted `.task-board` paths, preserved).
- Board: STORY-260908-3mcpz5 = integrating, TASK-260908-qblycn = integrating (unchanged after the attempt).
- Scoped config `/Users/iv/Developer/ReluxWorks/curator-agent-launcher/.temp/launcher-migration/task-board.config.json` declares `spawn.worktree_isolation.board_repository.root=/Users/iv/Developer/ReluxWorks/curator` and validation `make check` (confirmed by reading the file).

## Command (run from control root /Users/iv/Developer/ReluxWorks/.worktrees/launcher-control, inherited TASK_BOARD_CONFIG)
```
task-board --no-update-check --board-dir /Users/iv/Developer/ReluxWorks/curator/.task-board worktree complete STORY-260908-3mcpz5 --cr TASK-260908-qblycn --revision 1 --commit-time 2026-09-07T21:30:00+03:00 --json
```
Process exit status: **0** (stdout empty; the error JSON went to stderr). Note: a refusal reporting exit 0 is itself a tool-contract anomaly worth a reusable-tool fix, but it is not what blocks closure.

stderr (verbatim):
```
{"error":{"code":"VALIDATION_ERROR","message":"worktree_protected_authority_indeterminate: the repository does not have exactly one authorized fetch remote (control_root=/Users/iv/Developer/ReluxWorks/curator, remedy=repair repository binding or fresh remote HEAD evidence and retry; no local or caller fallback is authorized, remote_count=2, remotes=origin,origin-https)","details":{"code":"worktree_protected_authority_indeterminate","control_root":"/Users/iv/Developer/ReluxWorks/curator","remedy":"repair repository binding or fresh remote HEAD evidence and retry; no local or caller fallback is authorized","remote_count":"2","remotes":"origin,origin-https"}}}
```

## Root cause (evidence)
The board-owning Curator checkout has two fetch remotes pointing at the same URL:
```
origin        https://github.com/relux-works/curator.git (fetch/push)
origin-https  https://github.com/relux-works/curator.git (fetch/push)
```
`worktree complete` resolves the board repository's fresh protected default and requires exactly one authorized fetch remote; with two it refuses as indeterminate before the transaction is created. `worktree transaction show STORY-260908-3mcpz5` -> `{"transaction": null}`: nothing was written, nothing to dispose. No board, lease, workspace, CR or transaction record was touched; no push, commit, tag, LOGBOOK or ax action occurred.

The launcher control root itself is fine (workspace status: authority_source=fresh_protected_authority, upstream_remote=origin, upstream_oid=484933b). The refusal is specific to the board repository binding.

## Concrete next repair (operator / parent, not performed here)
Option A (recommended, minimal): remove the duplicate remote from the Curator checkout, then re-run the same command:
```
git -C /Users/iv/Developer/ReluxWorks/curator remote remove origin-https
```
This is a git-config change on the user's checkout, not board or working-tree state; it was NOT executed in this run because the brief limits this run to the complete command and forbids changing foreign state.
Option B (if the duplicate remote is intentional): make the binding explicit for the board repository (a `board_repository` remote-name binding equivalent to `TASK_BOARD_REPOSITORY_REMOTE_NAME=origin` for the control root). Whether the installed 416693dd config schema supports that for `board_repository` was not verified in this run; if unsupported it is the smallest reusable-tool gap.

After either fix: pull --ff-only again, re-run the complete command, then verify Story/task done, the transaction shows board_published, and the "Record STORY-260908-3mcpz5 board state" commit is signed and contained in origin/main.

## Files
- complete-01.json (empty stdout), complete-01.stderr (error above) under control-root `.temp/STORY-260908-3mcpz5/a0-complete/`.
