# TASK-260910-gocke2 — security remediation leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260910-1lf0m5. Normative spec: curator-spec v1.0.0-rc.13
(SPEC_PIN on curator main = 23435129); cite the clauses/vectors you implement. Your workspace starts on current trunk.
An earlier attempt (base from 2026-09-17..22) was snapshotted as refs/campaign/1lf0m5-full-20260927 and its workspace discarded — use it ONLY as a
reference (`git show refs/campaign/1lf0m5-full-20260927` / diff vs its parent); trunk moved a lot, never apply it blindly.

1. Implement through the production entry, fail-closed where the spec says so.
2. Drive the pinned rc.13 vectors for this surface; conformance-gaps.tsv rows owned by STORY-260910-1lf0m5 (or this task) that now pass LEAVE the ledger;
   rows whose first blocker is another not-yet-implemented surface are re-attributed to that surface's Story. Report before/after counts.
3. One mutant per rule — survive before, killed after (real exit codes).
4. No CHANGELOG/LOGBOOK edit (entry text in results). Focused bounded runs (host memory); hosted gate is the arbiter; artifacts only in $TMPDIR.
Attach results, check DoD, `task-board handoff TASK-260910-gocke2 --role developer`. A write-boundary `policy warn` block is a warning.
