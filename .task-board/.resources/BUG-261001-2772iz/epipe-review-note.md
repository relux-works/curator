# Review note — BUG-261001-2772iz rev2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Security-relevant review: the askpass secret pipe. The CR is rev2, base 54d4afac, 6 paths, gate green. Review it against `epipe-brief.md` and `epipe-gate-note.md`.

Check:
1. **Deterministic repro.** The new test forces the case where the child exits before the write, without -race and without timing luck. It FAILS on base 54d4afac with the broken-pipe error and passes on the candidate. Run both with real exit codes.
2. **Security invariants.**
   - Refusal outcomes are unchanged.
   - The secret is never written elsewhere and never logged.
   - A closed pipe is NEVER treated as successful auth.
   - A genuine write failure while the child is reading still surfaces as an error. Show the row that proves this.
   - Windows: the `_windows_test` file is meaningful, not a skip-only stub.
3. **Happy path on all OSes.** The rev1 Linux regression must be fixed: `go test ./internal/buildrepo -run HTTPSCredentialBroker -count=20` exits 0. Also check the crossconformance `TestDraftSourcesBrokerAskpassDispatch` rows with `-race -count=50` on darwin.
4. **Mutant.** Revert the EPIPE classification; the deterministic test must fail.
5. **Hygiene.** One CHANGELOG line only; no LOGBOOK; never spell any employer name.

accept_cr, or changes requested with file:line.
