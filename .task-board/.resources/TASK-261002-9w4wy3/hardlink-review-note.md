# Review note — TASK-261002-9w4wy3 rev1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, R138); security-relevant. Curator CR, base 2cb29dac, 13 paths, gate green on all lanes including Windows. Review it against `hardlink-brief.md` and STORY-260925-1v7pvn.

Verify, with real exit codes:
1. **Production resolver.** Both `windows-exec-noncomponent-store-hardlinks` and `windows-exec-unowned-file-hardlinks` refuse in the production resolver, not only in a helper. The component-store exception and the uncaptured-SystemRoot rejection are preserved. All 8 family cases are counted through the production entry.
2. **Native APIs.** The Windows code uses real platform APIs for link enumeration and owner identity, and errors fail closed. A read or API failure must never become "accept". The non-Windows seam must not let tests pass for the wrong reason.
3. **Ledgers.** Exactly 2 gap rows are removed from `.github/ci/conformance-gaps.tsv`, with exact counts. Any `platform-cases.tsv` change must be justified, with no new blanket skip. `docs/ci-gates.md` must be accurate.
4. **Mutants.** Kill at least 2 yourself: drop each new check.
5. **Hygiene.** One CHANGELOG line; no LOGBOOK; never spell any employer name.

The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down.

accept_cr, or changes requested with file:line.
