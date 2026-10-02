# THE ONLY CURRENT INSTRUCTION — TASK-261002-3so4n6: prepare curator-agent-launcher v0.2.0 (NO tag)

The operator approved v0.2.0. Start from public main 1ac7eaf; tag v0.1.1 already exists on it and must never move. Per the readiness checklist (curator `.research/261002_rc14_rc3_readiness.md`, launcher section):

1. Set `cmd/curator-run/main.go` buildVersion to "0.2.0". Keep specVersion "0.5.0-draft" unless the launcher contract itself changed (state which in the results). Update the version goldens.
2. CHANGELOG: a dated "0.2.0" entry built from Unreleased. It covers:
   - the muse mapping and v3 fragments;
   - the interactive Muse root session on agents-management v0.5.37;
   - the CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION env (curator#102);
   - shared typed context construction.

   Be accurate against `git log v0.1.0..origin/main`.
3. README:
   - the install line `go install github.com/relux-works/curator-agent-launcher/cmd/curator-run@v0.2.0`;
   - the supported-environments sentence must include Muse;
   - the compatible curator version is "curator v0.15.0-rc.3 or later, once published".
4. Run `go test -p 1 ./...` (bounded) and `GOOS=windows go vet ./...`, comparing the latter with baseline. Record real exit codes.

No tag. Never edit LOGBOOK.md. Never spell any employer name. The host has syspolicyd exec stalls: wait while syspolicyd is down.

Update the results, then run `task-board handoff TASK-261002-3so4n6 --role developer`, then END YOUR TURN.

## Decision (binding, 2026-10-02 ~05:40Z): host stall is not a blocker for handoff
The local full-suite stall comes from the known syspolicyd crash-loop on this host, not from the change. The CR validation suite runs the full matrix on hosted GitHub runners, and that hosted gate is the arbiter.

Do:
- Record exactly which local packages ran (with exit codes) and which stalled.
- Do not retry the stalled helper more than once, and only while syspolicyd shows "running" with a stable crash count.
- Then hand off for review. Do not set the task to blocked for this reason.
