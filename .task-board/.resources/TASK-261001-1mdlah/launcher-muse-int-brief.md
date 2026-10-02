# THE ONLY CURRENT INSTRUCTION — TASK-261001-1mdlah: curator run muse, interactive (curator-agent-launcher)

## Context
Already landed:
- launcher ee66c107: the muse mapping and the launch-env-fragment-v3 reader. The interactive plan.Build refusal is pinned there as a TESTED BOUND that flips once a plugin declaring interactive is pinned.
- curator e4a6a8d5: emits v3.
- curator-spec e41c561b: the muse adapter.

agents-management v0.5.34 is tagged. It carries skill-agents-management PR #50, TASK-260930-yiw33v:
- the Muse plugin declares LaunchModeInteractive with typed native/yolo per Decision 0018;
- native passes no posture flag;
- yolo emits --yolo once;
- a per-release policy row fails closed for unlisted releases;
- exec/dry-run argv stays byte-identical.

## Do
1. Bump agents-management to v0.5.34 (go.mod/go.sum) and update every doc that names the pin (README, SPEC).
2. Flip the bound. `curator run muse` with a canned v3 fragment and a FAKE muse binary must now:
   - build an interactive plan;
   - set the four XDG vars and NOT set HOME;
   - pass no posture flag for `--permissions native`;
   - pass exactly one `--yolo` for the yolo permission;
   - refuse a muse release not listed in the policy row (fail closed).

   Assert exact argv and env. Mutants: the HOME-set mutant fails; a double --yolo fails.
3. The Claude goldens change because the Claude plugin adds `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false`. Update them and name that in the CHANGELOG; it delivers the launcher half of curator#102. Any OTHER golden or argv change versus v0.5.22 must be listed and justified.
4. Run `go test -p 1 ./...` and `GOOS=windows go vet ./...`; compare the Windows output with baseline. Record real exit codes. NEVER start a real muse/claude/codex session or log in.

Never spell any employer name. No LOGBOOK. CHANGELOG under Unreleased.

## Handoff
Update the results, then run `task-board handoff TASK-261001-1mdlah --role developer`, then END YOUR TURN.

## Update (binding, 2026-10-01 ~20:10Z): blocker resolved upstream; pin v0.5.37
skill-agents-management PR #53 (R1c) adds the Muse ToolReleaseProber and a policy row for Muse **1.4.2**. It is tagged **v0.5.37** (peeled 8ee0150).

Do:
- Re-pin to v0.5.37 instead of v0.5.34. Keep all the WIP already in the worktree.
- In the strict root rows, make the fake muse report **1.4.2** (the listed release). Its `--version` output must match exactly what the new prober parses: read the upstream prober and its tests.
- Make **1.4.1** your "unlisted release refused" row (fail closed).

Every other requirement above stands:
- native passes no posture flag; yolo passes exactly one `--yolo`;
- XDG only, HOME never set;
- mutants;
- the #102 goldens;
- the Windows vet comparison.

Re-hand off when green.

## Decision (binding, 2026-10-01 ~23:10Z): unlisted-release row
v0.5.37 admits Muse 1.4.1 as well as 1.4.2, so my earlier "1.4.1 refused" instruction was wrong.

The fail-closed row must use a release that v0.5.37's policy does NOT list:
- 1.4.0 is fine if the policy lacks it; otherwise use a clearly synthetic one such as 0.0.1.
- Read the policy rows and cite them in the results.
- Keep 1.4.2 as the admitted interactive row. A 1.4.1 admission row is optional.

Hygiene:
- `TASK-261001-1mdlah_recovery-results.md` sits in the worktree ROOT. Remove it from the worktree; it must not be in the CR. Its content belongs in the task results resource.

The host has syspolicyd exec stalls. Before a long command, check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"`; if it is not running, wait instead of retrying. Run targeted packages with `-p 1` rather than a stalled full `./...`, and say exactly what ran.

Then hand off.
