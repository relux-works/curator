# Accepted leaf checkpoint evidence

CR-TASK-260908-25z3wj-1 revision 1 was checkpointed by the supported CLI from the frozen launcher-control root with inherited configuration. The leaf remains integrating; CR state is checkpointed. No generic handoff or manual commit was used.

Command: `task-board --no-update-check --board-dir /Users/iv/Developer/ReluxWorks/curator/.task-board worktree checkpoint TASK-260908-25z3wj --json`
Exit code: 0. Result: already_checkpointed=false, skipped=false, repository_delta=present.

- Checkpoint: e6827e350485c55c743c4cfcfc80ad6aeb4ed831
- Parent: 13b28c9a8916464e7253551808ae9969d6aa0186
- Accepted and checkpoint tree: 5b154a523b71ec5daabb97a55e81d9af1d3a5c86
- Real index (`git write-tree`): 5b154a523b71ec5daabb97a55e81d9af1d3a5c86
- Branch: refs/heads/task-board/story/STORY-260908-1wxjbs
- Author: Ivan Oparin <ivan@relux.works>

## Direct verification

| Command | Exit | Evidence |
| --- | ---: | --- |
| `git verify-commit HEAD` | 0 | Good git ED25519 signature for ivan@relux.works; SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0 |
| `git diff --exit-code` | 0 | No unstaged delta |
| `git diff --cached --exit-code` | 0 | No staged delta |
| `task-board q 'get(TASK-260908-25z3wj) { id status }'` | 0 | integrating |
| `task-board ... worktree status STORY-260908-1wxjbs --json` | 0 | Record checkpoint and branch tip both e6827e35; dirty=false; registered/present=true; CR checkpointed |
| `git merge-base --is-ancestor e6827e350485c55c743c4cfcfc80ad6aeb4ed831 refs/heads/main` | 1 | Expected negative: checkpoint is not landed in local main (18aeaed9af7dc5ffbe6cc79a4731a852fbb716da) |

The workspace CR index_tree_oid is historical candidate-capture metadata (d6457898); the refreshed real index equals the checkpoint tree. No record edits were made.

An initial read attempted unsupported `task(...)` and received an unknown-operation parse error. It was corrected to the supported `get(...)` projection. That failed read was in a multi-command inspection, so its standalone exit was not captured; no validation claim relies on it.

No tests, builds, mutants, source edits, installs, pipeline changes, SPEC changes, dependency changes, home writes, or delivery to main were performed in this integration run. Prior accepted producer/reviewer evidence is retained, not represented as rerun here. TASK-260909-2vy977 retains the real module Lineup and production pipeline/diagnostic scope. The parent owns the next leaf and final Story delivery.

Raw checkpoint JSON follows:
{
  "already_checkpointed": false,
  "branch_ref": "refs/heads/task-board/story/STORY-260908-1wxjbs",
  "commit_oid": "e6827e350485c55c743c4cfcfc80ad6aeb4ed831",
  "cr_id": "CR-TASK-260908-25z3wj-1",
  "id": "TASK-260908-25z3wj",
  "repository_delta": "present",
  "revision": 1,
  "skip_reason": "",
  "skipped": false,
  "status": "integrating"
}
