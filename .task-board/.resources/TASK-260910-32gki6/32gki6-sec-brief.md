# TASK-260910-32gki6 — security remediation leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260910-148pj1. The normative spec text is in
curator-spec v1.0.0-rc.13 (SPEC_PIN on curator main = tag commit 23435129): find the clauses and vectors this leaf implements and cite them.
1. Implement the rule(s) in curator through the production entry (CLI/install/resolve paths), fail-closed where the spec says so.
2. Drive the pinned-suite (rc.13) vectors for this surface through production-entry rows; every conformance-gaps.tsv row owned by STORY-260910-148pj1 (or
   this task) that now passes must LEAVE the ledger (the ratchet rejects passing gaps); report before/after counts.
3. Mutants: one per rule (e.g. the check removed / the error softened to a warning) — survive before, killed after (real exit codes).
4. No CHANGELOG/LOGBOOK edit (entry text in results under "## CHANGELOG entry (for release prep)"). Focused bounded runs (host memory);
   the hosted gate is the arbiter; artifacts only in $TMPDIR.
Attach results (rules → rows → vectors table), check DoD, `task-board handoff TASK-260910-32gki6 --role developer`, then END YOUR TURN (the runner publishes the CR and runs the gate).

Focus (S5, profile-store protected boundary): the manager's env resolve path verifies the environments root, profile store root, lock file, markers and every store entry the lock names with the five boundary checks (ownership, private mutation permissions / owner-only DACL, containment, regular file types, lstat link safety) plus pin-hash recomputation, with the two failure classes (enclosing boundary → refuse all; entry → environment_store_untrusted + rebuild from revalidated snapshot; unreadable lock never rebuilt) — environments.md §4 'protected state' (read it; E6 TASK-260916-yvxbs1 already landed the path-source part and an internal/pathboundary helper — reuse it).
New manager-state reads via internal/stateread; managed writes via the E5 nofollow helpers. Write only inside your Story worktree.
