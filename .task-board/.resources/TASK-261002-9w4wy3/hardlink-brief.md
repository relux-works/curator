# THE ONLY CURRENT INSTRUCTION — TASK-261002-9w4wy3 Windows executable hard-link origin checks (curator; pre-rc.3 security fix)

Read STORY-260925-1v7pvn (description and AC). The curator production resolver for Windows System32 executables (`internal/scriptworker`) still ACCEPTS two published spec cases from the dcc7f015 executable-identity family:
- `windows-exec-noncomponent-store-hardlinks`
- `windows-exec-unowned-file-hardlinks`

Do:
1. Reject both cases at the production resolver. Keep:
   - the platform-owned component-store (WinSxS) hard-link exception;
   - the uncaptured-SystemRoot rejection.

   Use real platform APIs (file ID / link enumeration / owner SID) behind a seam testable on non-Windows. Native behaviour must be exercised on the hosted Windows lanes.
2. Tests reach the production resolver and count all eight published cases of the family. Remove exactly the two gap-ledger rows once they pass; exact counts.
3. Mutants: drop each check in turn → the corresponding case fails. Run `GOOS=windows go vet ./...` and compare it with baseline. Record real exit codes.
4. If a platform API makes one check impossible to implement truthfully, stop and record the exact constraint and evidence; do not fake it.

No LOGBOOK; one CHANGELOG line. Never spell any employer name. The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down.

Update the results, then run `task-board handoff TASK-261002-9w4wy3 --role developer`, then END YOUR TURN.
