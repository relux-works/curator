# Review note for TASK-260906-1xbrz6 revision 2 (orchestrator, binding)

Revision 2 answers your revision-1 verdict (`TASK-260906-1xbrz6_review-verdict-rev1.md`): F1 — the
four exclusion sentences are now pinned as whitespace-normalised EXACT sentences (predicate changes
like fail closed→fail open, outside→inside, subject swaps must fail); F2 — complete heading lines are
anchored and sections are delimited at the next same-or-higher heading; your four read-only
mutations are committed as negative tests. Recovery claims: the import sentence now says retry the
ACTIVATION (not the import, which stops with `profile_import_name_taken`), citing §9.1/§9.2/§9.6; the
global-add lock-publication ordering is NOT settled by §9.4, so the producer attached
`TASK-260906-1xbrz6_global-recovery-decision-packet.md` instead of asserting it.

Judge: (1) re-run your four mutations — each must now FAIL the function; add one of your own (e.g.
move the exact sentence into §9.7, or duplicate it); (2) the import-activation reasoning against
§9.1/§9.2/§9.6; (3) whether the decision packet is correctly raised (the question is real, the
alternatives are complete, the recommendation is sound) and that the text does not over-claim the
global recovery meanwhile; (4) the gate evidence for revision 2 is green. A correctly raised packet is
NOT a defect of this leaf — the orchestrator routes the decision.

Record exactly one verdict: `accept_cr(TASK-260906-1xbrz6, revision=2, evidence=<your outcome
resource>)` on ACCEPT, or changes-requested with file:line and reproduction. No LOGBOOK.md writes.
