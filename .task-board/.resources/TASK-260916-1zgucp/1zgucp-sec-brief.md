# TASK-260916-1zgucp — security remediation leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260916-ioemse. The normative spec text is in
curator-spec v1.0.0-rc.13 (SPEC_PIN on curator main = tag commit 23435129): find the clauses and vectors this leaf implements and cite them.
1. Implement the rule(s) in curator through the production entry (CLI/install/resolve paths), fail-closed where the spec says so.
2. Drive the pinned-suite (rc.13) vectors for this surface through production-entry rows; every conformance-gaps.tsv row owned by STORY-260916-ioemse (or
   this task) that now passes must LEAVE the ledger (the ratchet rejects passing gaps); report before/after counts.
3. Mutants: one per rule (e.g. the check removed / the error softened to a warning) — survive before, killed after (real exit codes).
4. No CHANGELOG/LOGBOOK edit (entry text in results under "## CHANGELOG entry (for release prep)"). Focused bounded runs (host memory);
   the hosted gate is the arbiter; artifacts only in $TMPDIR.
Attach results (rules → rows → vectors table), check DoD, `task-board handoff TASK-260916-1zgucp --role developer`. A write-boundary `policy warn` block is a warning.
