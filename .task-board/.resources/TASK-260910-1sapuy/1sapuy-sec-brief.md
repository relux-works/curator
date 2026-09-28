# TASK-260910-1sapuy — security remediation leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260910-2qmrb8. The normative spec text is in
curator-spec v1.0.0-rc.13 (SPEC_PIN on curator main = tag commit 23435129): find the clauses and vectors this leaf implements and cite them.
1. Implement the rule(s) in curator through the production entry (CLI/install/resolve paths), fail-closed where the spec says so.
2. Drive the pinned-suite (rc.13) vectors for this surface through production-entry rows; every conformance-gaps.tsv row owned by STORY-260910-2qmrb8 (or
   this task) that now passes must LEAVE the ledger (the ratchet rejects passing gaps); report before/after counts.
3. Mutants: one per rule (e.g. the check removed / the error softened to a warning) — survive before, killed after (real exit codes).
4. No CHANGELOG/LOGBOOK edit (entry text in results under "## CHANGELOG entry (for release prep)"). Focused bounded runs (host memory);
   the hosted gate is the arbiter; artifacts only in $TMPDIR.
Attach results (rules → rows → vectors table), check DoD, `task-board handoff TASK-260910-1sapuy --role developer`. A write-boundary `policy warn` block is a warning.

Focus (S1/S3, secure defaults for audit gates): the spec side landed (TASK-260910-2qtiho hardened-defaults profile, TASK-260918-2mglq0
S1/S3 rebase). Per rc.13 (manager §7.1 hardened posture / registry evidence clauses — find and cite them): an unreachable trusted registry
during install/update must be a prominent gate notice (or, under the hardened posture, a refusal if the spec says so) naming every artifact
resolved without registry evidence — not a routine warning. Implement through `curator install`/`update`, drive the rc.13 vectors, remove
owned gap rows. Handoff runs the hosted gate — WAIT for it; hand off only green; if red, fix and republish. Write only inside your Story worktree.
