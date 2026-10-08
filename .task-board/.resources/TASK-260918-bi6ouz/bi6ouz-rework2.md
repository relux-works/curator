# THE ONLY CURRENT INSTRUCTION — TASK-260918-bi6ouz rework 2: fix the one red guard test and republish (developer)
**What happened.** Revision 8 (your re-application after the converge) went to the hosted gate; every test job failed on ONE test, the same on all platforms:
`internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded` (state_read_guard_test.go:218):
"internal/envprofile/managed.go:foreignManagerHint tests not-exist after function-value.os.Lstat without a seam route or reviewed allowlist reason" (489/490 manager-owned reads covered).
Since revision 7 was written, manager-owned absence reads must go through the read-guard **seam** (or carry a reviewed allowlist reason). Your `foreignManagerHint` does `os.Lstat` through a function value and tests not-exist directly.
**Do, in the workspace as it is now (do NOT checkout, reset or converge):**
1. Read `internal/envprofile/state_read_guard_test.go` (how the audit classifies reads, what "seam route" means, where the allowlist lives) and find how the other absence reads in `managed.go` are routed through the seam today. Route `foreignManagerHint`'s Lstat and not-exist test through that seam the same way. Do NOT add an allowlist entry unless the seam genuinely cannot express it; if you must, the reason text has to explain why, in the reviewed form the file documents.
2. Run the guard test and your dotfile tests locally through the lock: `~/.local/bin/mini-build-lock run bi6ouz -- env GOFLAGS=-work go test ./internal/envprofile -run 'TestManagerOwnedAbsenceReadsAreGuarded|Dotfile|Takeover|Managed' -count=1 -timeout=10m` (R193; the guard test takes ~1 min). No cmd/curator locally (R194).
3. Results resource (plain text, no archives): what you changed and the local test tail.
Then `task-board handoff TASK-260918-bi6ouz --role developer` and END YOUR TURN. No LOGBOOK or CHANGELOG edits.
