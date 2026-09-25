# TASK-260922-3bbvrs — rework 2 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.
(`3bbvrs-brief.md` still defines the deliverable; everything the reviewer verified in `TASK-260922-3bbvrs_review-verdict-rev2*` holds.)

Blocking finding — the consumer list is incomplete. These loops render EVERY case of a published family without the
ledger or a count pin:
1. `internal/devsub/repository_test.go:43-57` — every file in `schema-cases/skillfile-dev-v2` (16 at rc.12).
2. `internal/buildrepo/admission_test.go:191` — every case in `fixtures/external-repository/raw-objects.json` (12).
3. `internal/buildrepo/admission_test.go:240` — every case in `fixtures/external-repository/lfs-pointers.json` (10).
4. Check `admission_test.go` ~269 (`pack-index.json`) and ~396 (`local-config-and-refs.json`) the same way.
Route each whole-family loop through `conformancecoverage.Run`/`RunOutcomes` with its pin in
`conformance-case-counts.tsv` (and `root-artifacts.tsv` if the gate requires it), or record a narrow reasoned exclusion
in the docs. Filtered single-case lookups (buildmeta, whitelist, skillcheck) stay as they are — say so in results with
the grep you used to find ALL loops over published case lists. One narrowing mutant on one newly routed family (delete a
case from the pinned count, or let a failing case through) → killed. Bounded local runs; gate scripts. Append
"Revision 3" to results, then `task-board handoff TASK-260922-3bbvrs --role developer`. A `run_wrote_outside_worktree …
policy warn` block is a warning — verify status `to-review`.
