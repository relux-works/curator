# THE ONLY CURRENT INSTRUCTION — BUG-260923-2afgyq marker reader cross-field validation (curator; pre-rc.3 security fix)

Read the element's description and AC. `marker.Read` admits five invalid install-marker-v4 external-repository cases:
- declared vs effective identity mismatch;
- local vs network identity-kind mismatch;
- SHA-1 vs SHA-256 effective-revision width mismatch.

They are listed as gap rows owned by this bug in the conformance gap ledger (`.github/ci/conformance-gaps.tsv`).

Do:
1. Add the cross-field checks in `internal/marker` (validV3Build / v4 path) at the production reader. Do NOT change released marker schemas.
2. All five published invalid cases now refuse. All currently valid marker fixtures still pass, for every marker version.
3. Remove exactly those five ledger rows, and only once their exact published cases pass; keep exact counts. Use real exit codes for `go test ./internal/marker ./internal/conformancecoverage -count=1` and the relevant cmd/curator rows.
4. Mutant: drop each new check in turn and show the matching case fails.

No LOGBOOK. One CHANGELOG line under Unreleased. Never spell any employer name. The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down.

Update the results, then run `task-board handoff BUG-260923-2afgyq --role developer`, then END YOUR TURN.
