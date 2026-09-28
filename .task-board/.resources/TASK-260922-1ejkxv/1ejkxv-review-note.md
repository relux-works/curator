# Review note — TASK-260922-1ejkxv rev1, F-S3 fleet-enforced isolated mode (orchestrator, binding)

Review CR revision 1 against `1ejkxv-brief.md` (+ `1ejkxv-rework-1.md`) and the results. Read-only; run anything in a
disposable clone (not git archive) of the candidate tree. The workspace was converged onto spec main `eadb1c0`
(1xbrz6) before the gate: confirm 1xbrz6's text/pins in the five shared files survived the merge intact.
1. `system-config-v2` admits the `isolated` lock direction with EXACTLY the manager §1 locked-list semantics `shared`
   has (same shape, precedence, no new knob); 0017 choice 5 (knob is the only mode field) still holds.
2. environments §12.2 states both directions; `shared` requested under an engaged `isolated` lock → named
   diagnostic (existing code reused or a new one registered in the diagnostics table); silence → locked direction;
   an already-provisioned shared passthrough when the lock engages fails closed (never silently migrates) and
   points at F-C2's explicit migration.
3. Schema cases: valid isolated, valid shared (unchanged), invalid both-directions, invalid unknown direction —
   each validated by `tools/validate.py`; generated manifest/index regenerate-check clean.
4. results names the curator follow-up leaf (manager-side enforcement) with the exact schema key and diagnostic.
5. Gate = the recipe lines run in bounded parts (rework-1 accepted that composition) — check the table covers all
   three `validate:` lines + regenerate-check with exit 0; plus the runtime validation log.
6. Your own mutant (disposable clone): e.g. drop `isolated` from the enum → which case fails.
Findings → changes requested with file:line; else accept_cr. No LOGBOOK.md.
