# TASK-261018-adogt7: publication retry blocker

## Publication retry (2026-10-19)

Status: **blocked before publication** by managed-worktree board projection synchronization. The earlier runner cleanup failure is not being reproduced or counted as passing in this run; the retry brief reports its external repair.

- No source, test or documentation file was changed. The full physical workspace tree is `cd1a903247725014af5726b39bee9b1e90482959`, identical to the last signed detached candidate `e57894d642993b9fce7ea23f7ccfdb1995c2ac38` (signature `G`). Story checkpoint remains `9c1605a74436d2448f759bd4b66ae6685237d83b`; the real index is unchanged and root README is untouched. The pre-existing CHANGELOG edit was left unchanged per the retry brief.
- `task-board worktree status STORY-261018-9i0hmn` exited 0 and reported stale revision 1 with 19 changed paths. The assigned deferred-convergence target is `b9caa20901ee859eec20e649940693033e6297d4`.
- `task-board worktree refresh-candidate TASK-261018-adogt7` exited **1**, `change_request_candidate_drift`: eight tracked board paths outside the producer file transaction differ from prepared checkpoint tree `09b478fc86f54e1596c3fb07efb9857c50b6c186`. Refusal happened before branch/index swap; no refreshed combined candidate was produced.
- Identity verification exited 0; `git diff --check` exited 0. Full local suites and hosted gate were **not run in this retry**: a refreshed combined candidate does not exist. Previously attached green local evidence remains bound only to its unchanged historical source/test/configuration/environment identity; no success on current trunk is claimed. No new post-turn-validation checklist item was added.
- The unchanged coverage map and mutant table below remain historical evidence: **9 of 9 AC rows driven**, with the AC 9 hosted-green requirement still unverified for a new candidate. No new tests or mutant runs are claimed. There is still no supplied surface table and no AC row is declared out of contract.

### Publication retry logbook / recovery boundary

The tracked board projection is outside the producer file transaction. The task forbids direct board-file edits, and the project-management skill likewise requires CLI-only board writes. The installed refresh guard explicitly refuses stale non-resource board paths; its resource reconciliation only repairs absent immutable resources, not activity/history. Inspection of the CLI exposed no supported board-projection synchronization operation. The source of that guard is `tools/board-cli/internal/changerequest/refresh_postcondition.go:verifyRefreshBoardReady` in the project-management tool source. The task's write boundary does not authorize changing or installing that separate shared tool repository.

Required owner input: the orchestration/tool owner must synchronize this Story's checkout board projection to the prepared checkpoint through a supported repair, preserving authoritative activity/history and the uncommitted product candidate, or supply a supported CLI recovery that performs that synchronization. Then rerun the normal producer refresh, inspect any real combination delta, run the affected checks and hand off for the configured hosted gate. Repairing the shared contract at its source is preferable to a one-off direct board rewrite. Bypassing the guard, hand-editing board history, committing/rebasing the managed Story branch, and attesting an old tree as current are not valid alternatives.

The refusal and candidate identity are attached separately as `TASK-261018-adogt7_publication-retry-blocker.md`. No handoff is admitted while required base convergence is refused. The task is routed blocked for the named ownership repair; review and integration remain unaccepted.

## Exact refusal (command exit 1)

```text
change_request_candidate_drift: candidate refresh would leave tracked board paths outside its file transaction stale; synchronize the board with the prepared checkpoint before retrying (paths=.task-board/.activity/EPIC-260905-29w4hn/events.ndjson, .task-board/.activity/STORY-261017-7eqh0t/events.ndjson, .task-board/.activity/TASK-261017-sxydvh/events.ndjson, .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-261017-7eqh0t_audit-reserved-tld-hosts/README.md, .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-261017-7eqh0t_audit-reserved-tld-hosts/TASK-261017-sxydvh_reserved-tld-hosts-not-network/README.md, .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-261017-7eqh0t_audit-reserved-tld-hosts/TASK-261017-sxydvh_reserved-tld-hosts-not-network/progress.md, .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-261017-7eqh0t_audit-reserved-tld-hosts/progress.md, .task-board/EPIC-260905-29w4hn_agent-environments-execution/progress.md, prepared_tree_oid=09b478fc86f54e1596c3fb07efb9857c50b6c186, workspace_tree_oid=cd1a903247725014af5726b39bee9b1e90482959)
```

## Identity verification (exit 0)

```json
{
  "checkpoint": "9c1605a74436d2448f759bd4b66ae6685237d83b",
  "current_tree": "cd1a903247725014af5726b39bee9b1e90482959",
  "previous_signed_candidate": "e57894d642993b9fce7ea23f7ccfdb1995c2ac38",
  "previous_candidate_tree": "cd1a903247725014af5726b39bee9b1e90482959",
  "candidate_signature": "G",
  "index_unchanged": true,
  "README_unchanged": true,
  "refresh_exit_code": 1,
  "tree_equals_previous_signed_candidate": true
}
```
