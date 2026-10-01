# THE ONLY CURRENT INSTRUCTION — BUG-261001-2772iz askpass pipe EPIPE on refusal paths (curator)

## Symptom
Hosted gate run 36827469593, lane Race (macos-latest). `internal/crossconformance TestDraftSourcesBrokerAskpassDispatch` fails in four subtests: foreign_user, bare_prompt, no_arguments and extra_argument. Each fails with:
`draftsources_broker_e2e_test.go:141: HTTPS broker secret transport: write |1: broken pipe`

All other lanes pass. The askpass secret-via-pipe change (STORY-260928-1t6bto, landed 29ebed12) is the likely origin. On a refusal path the askpass child exits without reading the pipe. The broker's secret write then races the child's exit; when the child wins, the write fails with EPIPE and surfaces as a transport error.

## Required
1. Find the broker/askpass pipe writer and the exact race. Make it deterministic with a test that forces the ordering (the child exits before the write) WITHOUT -race or timing luck. That test must fail on current main with the broken-pipe error; record its real exit code.
2. Fix it so that a refusal path yields the intended refusal outcome, as the existing subtests expect, never a transport error. Security constraints:
   - the secret is never written anywhere else and never logged;
   - the refusal decision is unchanged;
   - a genuine transport failure while the child IS reading must still be reported;
   - a closed pipe must never be treated as successful authentication.

   Prefer a design where the broker does not write the secret before the child has asked, or where it classifies EPIPE only when the child has already exited with its refusal status. Justify the choice in the results.
3. Run `go test -race ./internal/crossconformance -run TestDraftSourcesBrokerAskpassDispatch -count=50` on darwin plus the relevant broker/askpass packages, with real exit codes.
4. Mutant: revert the fix and show the deterministic test fails.

Do NOT touch LOGBOOK.md. CHANGELOG: one line under Unreleased if the section exists. Never spell any employer name.

## Handoff
Update the results, then run `task-board handoff BUG-261001-2772iz --role developer`, then END YOUR TURN.
