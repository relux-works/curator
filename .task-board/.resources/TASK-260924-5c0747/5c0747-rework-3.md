# TASK-260924-5c0747 — rc.2 prep revision 3 on trunk f02ba39e (THE ONLY CURRENT INSTRUCTION, with 5c0747-brief.md)

Revision 2 (gate green) was withdrawn: its base e8620502 predates 2gt5f6 (316438cc) and h4syhu (f02ba39e — Story 1a2i5a: the §8.4 class
guard 2as5sx + 187z6x + h4syhu). The old workspace could not converge (1 conflicting path) and was DISCARDED after a full snapshot:
refs/campaign/5c0747-rev2-full-20260926 (parent e8620502). Your new workspace starts on trunk f02ba39e.
1. `task-board m 'set_status(TASK-260924-5c0747, status=development)'`.
2. Re-apply ONLY the release-prep delta: `git diff 316438cc refs/campaign/5c0747-rev2-full-20260926 -- . ':!.task-board'` is your 14-path
   release prep relative to trunk-with-2gt5f6 (verify it is exactly those 14 paths), apply with `git apply --3way`, resolve the conflict(s)
   keeping trunk's h4syhu/2as5sx content plus your pin/corpus/CHANGELOG changes; list each conflict. Leave nothing staged.
3. CHANGELOG v0.15.0-rc.2 section: ADD the "## CHANGELOG entry (for release prep)" texts of TASK-260925-h4syhu, TASK-260907-2as5sx and
   TASK-260907-187z6x (and any other leaf landed after v0.15.0-rc.1 you had not yet included — re-check git log v0.15.0-rc.1..origin/main).
4. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` lists only release-prep paths (pins, corpus pin, gap/counts ledgers, CI,
   docs, CHANGELOG, the tests the pin move requires) — no production code.
5. Bounded runs of the conformance packages you touch; `bash .github/ci/gate-selftest.sh`. Real exit codes. Append "Revision 3 — onto f02ba39e",
   `resource update`, `task-board handoff TASK-260924-5c0747 --role developer`; stay in the turn. No LOGBOOK edit.
