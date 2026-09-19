# TASK-260918-fjl6v2 integration attempt 2 — refused: revalidation_unavailable, trunk unmoved

Run: RUN-260918-cd849a, 2026-09-19 ~01:33–01:48 +04 (Asia/Dubai local).
Story: STORY-260918-2yvd86, leaf TASK-260918-fjl6v2 revision 1 (accepted).
Instruction: TASK-260918-fjl6v2_integration-run-attempt2.md. Outcome per step 4–5: refusal quoted verbatim, stop, no repair.

## 1. Pre-flight (step 1) — all confirmed, exit 0

`task-board worktree obligations` (exit 0) contained the row:

```
TASK-260918-fjl6v2     1    accepted  checkpoint  22m      integrating  STORY-260918-2yvd86
```

`git -C /Users/administrator/Developer/ReluxWorks/curator/curator status --short` (exit 0):
only `.task-board/` modifications/untracked resources — no non-board changes.
`git log --oneline -1` (exit 0): `6f17377 Record board state: 1ll22r and 12lbww spec revisions landed, manager leaves created`.

## 2. Detached launch (step 2) — pid 59213

The managed shell in this session rejects `&` backgrounding, so the brief's
`nohup python3 -c 'os.setsid(); exec...' ... &` was executed in its equivalent
non-`&` form — a short-lived Python parent using `subprocess.Popen(...,
start_new_session=True)` (= setsid detachment) from the control root with the
run environment intact (`TASK_BOARD_RUN_ID=RUN-260918-cd849a`,
`TASK_BOARD_DIR=.../curator/.task-board`), stdout+stderr to
`/tmp/integrate-2yvd86.log`:

```
pid=59213
```

(launch command exit 0). No other board mutation was run while it lived.
First check (+15s): process running, log still empty.

## 3. Polls (step 3) — exited within ~4 minutes

Short `tail` + `ps` polls only. At ~+4min the process had exited; the log tail
already showed the refusal (full log in §4). Nothing was killed or re-run.

## 4. Refusal — FULL /tmp/integrate-2yvd86.log quoted verbatim

```
revalidation_unavailable: the validation suite could not be executed against the tree that would land; this is not a pass and not a rework signal, and integration is blocked and retryable
  log: 
  merged_tree: 094dfcc5c8edba97f0841a0ca074766d4819d5c3
```

(exit code of the integrate process itself was not captured — detached launch
returns only the pid; the outcome is a refusal, i.e. non-zero.)
This names NO transaction phase and NO rollback condition, so per the brief:
no repair, no reset, no manual commit, no push, no reconcile-trunk, no
`--rollback`. Note `merged_tree` is byte-identical to the tree the
predecessor reports as GitHub-gate GREEN (run 35390532501, all lanes).

## 5. Post-exit state (step 4) — trunk and board unmoved, no commits

`git log --oneline -4` (exit 0) — unchanged head:

```
6f17377 Record board state: 1ll22r and 12lbww spec revisions landed, manager leaves created
d91b11a Record STORY-260910-2xe3n2 board state
e857e50 Record STORY-260910-20sx61 board state: completed lane paths and the remaining wave-3 records
020080d Record STORY-260910-20sx61 board state
```

`git status --short` (exit 0): 39 lines, all under `.task-board/` — board-only,
no candidate or landing residue. No squash commit and no board commit exist
(`verify-commit` not applicable — nothing landed).
`worktree obligations` (exit 0, via pinned binary, see §6): row unchanged —
`TASK-260918-fjl6v2 1 accepted checkpoint … integrating STORY-260918-2yvd86`.
`q 'get(STORY-260918-2yvd86) { status }'` → `{"status":"integrating"}`.
`q 'get(TASK-260918-fjl6v2) { status }'` → `{"status":"integrating"}`.

## 6. Tool-chain finding: default `task-board` binary hangs + SIGKILL after the run

Immediately after the integrate process exited, EVERY new invocation of the
default `task-board` wrapper
(`/Users/administrator/.curator/global/bin/task-board` → exec of
`/Users/administrator/.curator/cache/build/go-v1/e77b151704d2a922f9afba8207265c32be4da0ddfe6cc8c3a37371d1ee1b6c67/bin/task-board`)
hung for minutes and was then SIGKILLed (shell reports `Killed: 9`, exit 137)
— 4 consecutive invocations (`worktree obligations`, `q` ×2, and
`--no-update-check q`), each killed after ~1–4 min. Observed live, one stuck
invocation was still the `/bin/sh` wrapper (never reached exec) 76s in. No
git locks exist, no integrate/remote-gate children linger, no board `.locks`
dir exists, memory pressure is normal (~1GB free, load ~4.7), and all
pre-existing task-board processes (spawn-runners, spawn waits) are unaffected.
The orchestrator-pinned binary
`/Users/administrator/.local/bin/task-board-main-6cb09a23-curatorlike
--no-update-check` answers the same board instantly (exit 0); all §5 board
reads and this outcome's attach used it with ambient `TASK_BOARD_DIR`.
Same 137 signature as the predecessor's `revalidation_failed … exit_status:
137`. Recommend the orchestrator/tool owner investigate the cached go-v1
binary (and whether `revalidation_unavailable` with an empty `log:` shares
the cause) before the next integrate attempt.

## 7. Routing

Integration NOT landed (refused, retryable, trunk unmoved, board at
`integrating`). Per the brief the orchestrator routes the retry/refresh —
this run stops here. On retry, the revalidation backend should be confirmed
available first; the landing tree `094dfcc5…` is already gate-green upstream.
