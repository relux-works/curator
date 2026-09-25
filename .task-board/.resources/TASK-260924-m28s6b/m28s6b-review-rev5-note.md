# Review note — TASK-260924-m28s6b lock replay, newest revision (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 1 was reviewed (only the docs-pin remedy keyword was requested); since then: rework 1 (remedy/docs/test agreement), refresh
onto 5b326aa3, Windows no-bindings fix (Revision 3: portable file:// fixture URLs), refresh onto a48f584c (Revision 4) and a successor's
Revision 5 "Windows no-bindings redirect via repo-local config". Review the NEWEST revision against `m28s6b-brief.md` and the binding
operator addendum `skillfile-lock-replay-addendum-20260924.md`: missing snapshot re-materialized (git/repository fetch exactly the locked
revision, path reads current bytes); accept only when identity and content_sha256 equal the lock, else `source_snapshot_changed`;
`source_snapshot_unavailable` only for an unreachable source; lock byte-identical, no re-resolution. Judge the Windows changes: fixture-
only (as claimed) or a production behaviour change — if production, is it correct for a real Windows user? The Story also carries the
checkpointed 1aa9wb default-on (already accepted) — CHANGELOG.md must equal trunk, both entries in results. Hosted gate green on all
lanes (validation log). Kill at least one mutant of your choice in the replay acceptance rule (bounded, focused test only — host memory
is tight). accept_cr or changes requested with file:line. No LOGBOOK.md.
