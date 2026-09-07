# TASK-260908-qblycn integration run results (CR-TASK-260908-qblycn-1 rev 1)

Run: RUN-260907-6f5ac9, role researcher (analyst), story STORY-260908-3mcpz5.
Date (UTC): 2026-09-07T23:48Z. No research repeated; A0 findings live in the existing outcome resources.

## Outcome

The bound integration command was invoked once with correct operands and was refused
before any transaction opened. Nothing moved: story and task remain `integrating`,
no transaction record exists, trunk and control-root HEAD unchanged at 484933b, Curator
main unchanged at 04550e2 with its 98 dirty paths (board state + LOGBOOK) intact.

## Exact command and refusal

Control root cwd: /Users/iv/Developer/ReluxWorks/.worktrees/launcher-control (detached HEAD 484933b,
linked worktree; git-common-dir = /Users/iv/Developer/ReluxWorks/curator-agent-launcher/.git).
TASK_BOARD_DIR=/Users/iv/Developer/ReluxWorks/curator/.task-board.

```
task-board worktree integrate STORY-260908-3mcpz5 --cr TASK-260908-qblycn --revision 1 --json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "integration_indeterminate: /Users/iv/Developer/ReluxWorks/curator/.task-board is outside the control root\n  control_root: /Users/iv/Developer/ReluxWorks/.worktrees/launcher-control\n  path: /Users/iv/Developer/ReluxWorks/curator/.task-board",
    "details": {"code":"integration_indeterminate","control_root":"/Users/iv/Developer/ReluxWorks/.worktrees/launcher-control","path":"/Users/iv/Developer/ReluxWorks/curator/.task-board"}
  }
}
exit=0
task-board worktree transaction show STORY-260908-3mcpz5 --json  ->  {"transaction": null}
```

A prior attempt passed `--cr CR-TASK-260908-qblycn-1` and got `change_request_invalid_record`;
that was an operand usage error on my side (the flag takes the element ID the CR is attached
to), not a refusal of the integration path. It also opened no transaction. Logs:
execution worktree `.temp/TASK-260908-qblycn/integrate-01.log` and `integrate-02.log`.

Transaction disposition: none needed. The refusal happens before the transaction opens, so
`worktree transaction resolve/discharge` has nothing to act on and was not called.

## Where the refusal comes from (source read, board-cli 074c249d)

Source: /Users/iv/Developer/ReluxWorks/skill-project-management/tools/board-cli

- `internal/integration/integrate.go:130` acquires `<control-root>/.temp/integration/repository.lock`,
  then `:137` calls `NewBoardLayout(req.ControlRoot, req.BoardDir, ...)`.
- `internal/integration/paths.go:74` calls `repoRelative(controlRoot, boardDir)`; `paths.go:346-349`
  refuses with `integration_indeterminate` when the board dir is not under the control root.
- `cmd/worktree_integrate.go:246-247` binds `ControlRoot` to the repo control root and
  `BoardDir` to the resolved store dir (here the external TASK_BOARD_DIR).

This is by design, not a bug in the gate. The contract
`.research/260721_worktree-isolation-and-integration.md` states the control root "owns ...
the authoritative local board" (line 94), that `.task-board/**` is tracked in the same repository
(line 101), and that step 7 commits the board delta from the control root's own working tree
(lines 474-480). A board that lives in a different repository has no repo-relative board prefix,
so admission and the board-only commit cannot be computed.

## Second, independent blocker that would fire next

Even with an in-repo board, the §6.1 checkout preconditions (contract lines 340-352) require the
control root to be the repository's main working tree with attached HEAD on trunk. The current
control root is a linked worktree in detached HEAD, which maps to `integration_checkout_not_main`
and `integration_checkout_detached`. The launcher main working tree is
/Users/iv/Developer/ReluxWorks/curator-agent-launcher on `main` at 484933b.

## Confirmed facts (not inferred)

| Fact | Value |
| --- | --- |
| Story / task status after run | integrating / integrating |
| Transaction record | none |
| CR | CR-TASK-260908-qblycn-1 rev1, state accepted, kind story_final, repository_delta empty, producer researcher/analyst |
| Workspace | WS-38567b6935d9, branch tip = checkpoint = base = 484933b, clean, lease held by RUN-260907-6f5ac9 |
| Board-cli source rev | 074c249d (clean tree) |
| Curator main | 04550e2, 98 dirty paths preserved, LOGBOOK.md untouched by this run |

Unknown (not tested): whether `worktree checkpoint TASK-260908-qblycn` has the same control-root
dependency. It is not the bound command for a story_final CR and was not invoked.

## Concrete next action for the parent

The split-root shape (code in curator-agent-launcher, board in curator, control root a detached
linked worktree) is outside the supported contract of `worktree integrate`. Two legitimate routes:

1. Reusable-tool source change in skill-project-management board-cli: allow an external
   authoritative board by giving the board-state commit (step 7) its own repository root
   (the board's git root) instead of deriving `BoardPrefix` from the control root at
   `paths.go:74`, and by running the §6.1 checkout gates against the launcher main working tree.
   Smallest gap: `NewBoardLayout` and the step-7 commit assume one repository; that assumption
   needs to become an explicit `board_repository_root`. This is a design change to §2/§6.3, not a
   one-line fix, and must go through that repo's own board.
2. Operational route without a tool change: run the integration from a control root that is the
   launcher's main working tree on `main` with the board inside that repository. That does not
   match the operator requirement of a board-state commit in Curator main, so route 1 is the one
   that satisfies the stated requirement.

Recommendation: route 1, filed as a board-cli story; keep STORY-260908-3mcpz5 at `integrating`
until then. Do not hand-edit CR/transaction/workspace records or force status.
