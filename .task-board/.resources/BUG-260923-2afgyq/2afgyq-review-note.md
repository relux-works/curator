# Review note — BUG-260923-2afgyq rev1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, R138); security-relevant. Curator CR, base 2cb29dac, 4 paths, gate green. Review it against `2afgyq-brief.md`.

Verify, with real exit codes:
1. **Refusals.** The five published invalid install-marker-v4 cases now refuse at the production `marker.Read`:
   - declared/effective identity mismatch;
   - local/network identity-kind mismatch;
   - SHA-1/SHA-256 width mismatch.
2. **No regression.** Every currently valid marker fixture still reads, across all marker versions; run the marker and install packages.
3. **Ledger.** Exactly 5 gap rows are removed, with exact counts; `go test ./internal/conformancecoverage -count=1` passes.
4. **Mutants.** Drop each new check and confirm the matching case fails. Kill at least 2 yourself.
5. **Hygiene.** One CHANGELOG line; no LOGBOOK; never spell any employer name.

accept_cr, or changes requested with file:line.
