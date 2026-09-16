# Decision packet — TASK-260910-3du5nd (advanced transport-neutral endpoint mappings)

Constraint: the core transport-identity amendment (logical repository identity + machine transport policy) was accepted and landed through STORY-260910-8fv3s5 (curator-spec main a4fcaf0..a68854d lineage, protocol/repository-transport.md). What remains in this task is the ADVANCED part: custom ports, mirrors and endpoint aliases, listed in UNRESOLVED_QUESTIONS.md of the spec. The user instructed that this stays a separate future amendment and that no normative edits or implementation are authorized without a product decision.

Evidence: task notes (user approval scoped to the bounded core amendment only); spec UNRESOLVED_QUESTIONS.md; TASK-260910-16vtxi research resource.

Attempts: none beyond research — implementation is explicitly not authorized.

Alternatives:
A. Close this task now as "recorded" (the request and the residual questions are preserved in the spec's UNRESOLVED_QUESTIONS.md) and reopen a fresh task when the advanced mapping is scheduled. Lets STORY-260908-2haegq (B7 proposals) close.
B. Keep it open in backlog under B7 and file a fourth spec-gap proposal draft (ports/mirrors/aliases) as a decision document, then close the task on that draft.
C. Authorize the advanced amendment now (normative spec edit + implementation follow-up).

Recommendation: A (bookkeeping closure, nothing is lost) — or B if the operator wants the draft filed alongside the other three proposals.

Decision needed from the operator: A, B or C.
