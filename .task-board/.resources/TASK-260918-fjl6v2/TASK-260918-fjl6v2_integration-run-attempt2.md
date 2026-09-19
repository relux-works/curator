# Integration run (attempt 2, detached gate) — STORY-260918-2yvd86, final leaf TASK-260918-fjl6v2 revision 1 accepted

You are the tracked developer run bound to the accepted revision 1 of
`TASK-260918-fjl6v2`, the only leaf of story `STORY-260918-2yvd86` (the rc.12
conformance-pin promotion union, accepted by the independent reviewer in
`TASK-260918-fjl6v2_review-verdict-rev1.md`). Your predecessor
RUN-260918-df2574 ran `task-board worktree integrate` twice; both attempts
ended with `revalidation_failed … exit_status: 137` ONLY because the
`sleep` children of `scripts/remote-gate.sh`'s poll loop were SIGKILLed by
the harness's managed-bash session while the command ran in the foreground
for 30+ minutes (`TASK-260918-fjl6v2_integration.md`). The GitHub gate run
on the exact landing tree `094dfcc5c8edba97f0841a0ca074766d4819d5c3` was
GREEN (run 35390532501, all lanes). Trunk and board are unchanged; no
transaction phase was entered.

Run the integration DETACHED from your managed shell so nothing can kill its
children, and poll it with SHORT commands. Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm
   `TASK-260918-fjl6v2  1  accepted  checkpoint`; `git -C /Users/administrator/Developer/ReluxWorks/curator/curator status --short`
   must show no non-board changes; `git log --oneline -1` (current trunk).
2. Start the integration in its own session (macOS has no `setsid`; use
   Python's `os.setsid`) from the control root
   `/Users/administrator/Developer/ReluxWorks/curator/curator`, with your
   run's environment intact (do not unset or override `TASK_BOARD_RUN_ID`
   or the board variables):
   ```
   cd /Users/administrator/Developer/ReluxWorks/curator/curator
   nohup python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' \
     task-board worktree integrate STORY-260918-2yvd86 --cr TASK-260918-fjl6v2 --revision 1 \
     > /tmp/integrate-2yvd86.log 2>&1 &
   echo "pid=$!"
   ```
   Quote the pid. Do NOT run any other board mutation while it runs.
3. Poll: every 3–4 minutes run ONE short command (never a long sleep):
   `tail -n 5 /tmp/integrate-2yvd86.log; ps -p <pid> >/dev/null && echo running || echo exited`.
   The remote gate takes 35–50 minutes (Windows lane slowest). Keep polling
   until the process exits. Do not kill it, do not re-run it while it is
   alive, do not touch the worktree or the control root.
4. When it exits: quote the FULL `/tmp/integrate-2yvd86.log` (including
   `post_landing_steps` on success, or the typed error), then
   `git -C <control root> log --oneline -4`, `git status --short`,
   `git verify-commit HEAD~1` and `git verify-commit HEAD` (on success: the
   signed squash + the board-only commit), `task-board worktree obligations`,
   `task-board q 'get(STORY-260918-2yvd86) { status }'`. On success the story
   is `done`. If it exits 1 with `integration_indeterminate` AFTER the two
   commits exist (a known post-commit bookkeeping quirk), report that as
   landed-with-quirk with the commit ids. On any refusal quote it verbatim
   and stop: no repair, no reset, no manual commit, no push, no
   `reconcile-trunk`, no `--rollback` unless the output says the transaction
   is at phase `prepared` with trunk unmoved AND names a cause a re-run
   cannot fix.
5. Attach `TASK-260918-fjl6v2_integration-attempt2.md` (task outcome) with
   the transcripts and the commit ids, then exit.
