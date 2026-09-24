# Review note — TASK-260916-2ok97n R5, CR revision 7 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION). FINAL leaf of EPIC-260822.

First full content review of R5. Revisions 1-5 failed gates (CRLF ledger; Windows cmd.exe resolution; unresolved-exec semantics;
Windows build-worker Job Object binding; ledger row). Revision 6 was green on every hosted lane incl. Windows (run 35950409740) but carried stray CI artifacts (F1); revision 7 removed exactly those paths — first confirm rev6→rev7 differs ONLY by removing test/, ledger/ and the two root TASK-*.md files, then do the FULL content review below (items 3-6 were not done last cycle).
Review against `2ok97n-brief.md` (scope + rulings R1..), the rework briefs (rework-2..5), the Story AC and the results; read-only,
disposable clone of the candidate tree (base 1511b345). Source of truth: curator-spec protocol/core.md enforced script commands,
profiles/manager.md, `script-host-execution-policy.json` (rc.12 pin).
1. All 12 vector top-level keys consumed by production-entry rows (ratio 33/33 behavioural cases + 11 controls), no "helper-level only".
2. Ledger: every real launch/control/evidence/audit test registered with the right must/skip per lane; no new skip class; the
   CRLF-safe ledger reader.
3. Windows: declared `cmd.exe` resolves via the manager's default search list; hard-link allowance bounded EXACTLY (default list,
   manager-captured SystemRoot canonical System32, target below it) — this matches the accepted spec erratum TASK-260924-mcmova
   (PR curator-spec #88); unresolved declared exec → absent + REPORTED, never refused (core.md ~329-337); caller PATH never reached.
4. go-v1 build worker Windows evidence bound to the manager's private Job Object (duplicated query handle; missing/foreign handle
   refuse; kill-on-close preserved) — judge the change for race/handle-leak risks.
5. Stream model decision (R-e) documented or implemented as the brief required; narrowing negatives per remaining refusal gate with
   killed mutants — re-apply two yourself (e.g. farm drops a declared exec; caller-PATH fallback; Job Object membership check removed).
6. rose-air rows named for the landing run; no proof weakened or test deleted/skipped (R1).
accept_cr on revision 7, or changes requested with file:line. No LOGBOOK.md.
