# Review note — BUG-260924-5p8b0z, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `5p8b0z-brief.md` and the accepted spec erratum (curator-spec commit dcc7f015: core.md exec-capability bullet +
`executable_identity_cases`, 8 cases). Disposable clone.
1. The hard-link allowance now requires ALL erratum conditions conjunctively (Windows; declared exec; default search list; System32 derived
   from the MANAGER-CAPTURED SYSTEMROOT; target physically below it; extra links in that root's WinSxS). `windows-exec-uncaptured-systemroot-
   hardlinks` is rejected; the one accepting case still accepts; interpreters still reject every multiply-linked target.
2. All 8 executable_identity_cases driven at the production entry (vector bytes cited with sha256 if copied); the Windows lane in the hosted
   run shows them executing (read the gate's windows test evidence).
3. Cross-platform row via the GOOS/env seam; mutant (drop the captured-SystemRoot condition) killed — re-apply it yourself.
4. No CHANGELOG edit (entry in results); no stray files; validation log green.
accept_cr or changes requested with file:line. No LOGBOOK.md.
