# Review note — TASK-260924-2cp9w9 skillfile-sources residual clarifications, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `2cp9w9-brief.md` in a disposable clone of curator-spec (never the control root). Inputs: the gap matrix
`TASK-260924-3re9jo_results.md` group A and the cited verdicts (BUG-260920-2wyzde R1/R2, BUG-260920-2eg8nv §8, BUG-260920-3v7x6j R1/R3).
1. Each of the three clauses: the text settles the ambiguity, the reading is fail-closed (or the non-fail-closed choice is justified),
   it does not contradict other normative text (core.md, repository-transport, skillfile-sources elsewhere), and it is minimal (no
   status/namespace edits — another author is promoting these files out of "unreleased").
2. Each new vector distinguishes the clarified reading from the rejected one (name the mutant reading each case kills); the index and
   semantic-cases are consistent; regenerate-check and the three `validate:` recipe lines pass in your clone (bounded parts).
3. LOCKSTEP IMPACT (report, do not block on it): the spec "Implementations" check runs curator main against the spec candidate.
   Run curator's conformance suite at curator origin/main (a48f584c) against this candidate (disposable curator clone, spec pin pointed
   at the candidate tree) and list every NEW case curator fails, with the curator leaf that closes it (known: marker v5 local-arm nulls
   and declared_tag grammar → BUG 2v4v2m accepted, not yet landed; registry exact-grant rows → 10d3l1 accepted, not yet landed).
   Also confirm no new vector SECTION/family was added (a new section needs curator classification in lockstep).
4. No root TASK-*/BUG-* files, no test/ or ledger/ artefacts in the candidate. CHANGELOG entry under unreleased is fine in this repo.
accept_cr or changes requested with file:line. No LOGBOOK.md.
