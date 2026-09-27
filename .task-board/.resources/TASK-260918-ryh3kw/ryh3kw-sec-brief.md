# TASK-260918-ryh3kw — security remediation leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260916-1ll22r. Normative spec: curator-spec v1.0.0-rc.13
(SPEC_PIN on curator main = 23435129); cite the clauses/vectors you implement. Your workspace starts on current trunk.

MUCH OF THIS IS ALREADY ON MAIN: the shared classification helper (internal/stateread, cww1ov + h4syhu), the deny-by-default guard
(TASK-260907-2as5sx / TASK-260925-h4syhu, Story STORY-260906-1a2i5a), the read-site inventory. FIRST produce a table: each requirement of the
description → where it is on main (file:line / test) or MISSING. Implement only what is missing (e.g. the new environments diagnostics);
if nothing is missing, say so with evidence and hand off with an empty-delta justification.
1. Implement through the production entry, fail-closed where the spec says so.
2. Drive the pinned rc.13 vectors for this surface; conformance-gaps.tsv rows owned by STORY-260916-1ll22r (or this task) that now pass LEAVE the ledger;
   rows whose first blocker is another not-yet-implemented surface are re-attributed to that surface's Story. Report before/after counts.
3. One mutant per rule — survive before, killed after (real exit codes).
4. No CHANGELOG/LOGBOOK edit (entry text in results). Focused bounded runs (host memory); hosted gate is the arbiter; artifacts only in $TMPDIR.
Attach results, check DoD, `task-board handoff TASK-260918-ryh3kw --role developer`. A write-boundary `policy warn` block is a warning.
