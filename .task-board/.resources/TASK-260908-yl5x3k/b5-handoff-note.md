# B5 handoff note — TASK-260908-yl5x3k (2026-09-16, after PR72)

The onboarding itself is DONE and evidenced (RUN-260915-9c03fd: TASK-260908-yl5x3k_onboarding-evidence-rev2.md — umbrella v1.0.1 installed/activated, three managed homes current, real launches recorded). The previous handoff failed only because the landing gate script `scripts/remote-gate.sh` was not yet on curator main (exit 127). It is now (main 9213119). This spawn's refresh puts the Story worktree on current main.

Do NOT redo the onboarding. Do:
1. Confirm quickly (read-only): `curator profile list` shows relux-root-context-ivan current; `curator env status` shows claude_code, codex_cli, pi current+provisioned. Append a short "post-PR72 re-check" section with those outputs and exit codes to the existing evidence resource (update TASK-260908-yl5x3k_onboarding-evidence-rev2.md).
2. Verify the checklist is fully ticked (tick any unticked row whose evidence exists), then `task-board handoff TASK-260908-yl5x3k --role developer` in the FOREGROUND and wait for it: the handoff runs `sh scripts/remote-gate.sh`, which pushes a gate branch and waits for hosted CI (30–45 minutes). Do not background it; do not end your turn while it runs; a previous run kept a single shell call open for over two hours without being killed. If it fails, attach the runtime's validation log name and the exact failure and stop.

## Retry 2 (after converge onto main 18f0549)
The Story worktree now carries `scripts/remote-gate.sh`. The previous attempt also left a stray `TASK-260908-yl5x3k_onboarding-evidence.md` at the worktree root, which made the candidate carry a bogus repository delta; it has been removed. Keep the worktree CLEAN: write any scratch file under `/tmp` or the worktree's `.temp/` (gitignored) before attaching it as a resource. Then: confirm `git status --porcelain` is empty, and run `task-board handoff TASK-260908-yl5x3k --role developer` in the foreground and wait (30–45 minutes of hosted CI).
