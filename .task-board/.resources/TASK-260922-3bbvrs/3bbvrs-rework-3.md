# TASK-260922-3bbvrs — rework 3 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 3 fixed every revision-2 site. Remaining blocking finding (`TASK-260922-3bbvrs_review-verdict-rev3*` F1): the sweep
searched only `internal/`. Three whole-family loops in `cmd/` are unrouted:
- `cmd/curator/umbrella_conformance_test.go:81` — `family.Cases` over `vectors/umbrella-provider-resolution.json` (rc.12: 14);
- `cmd/curator/lifecycle_conformance_test.go:70` — `document.BootstrapCases` (manager-lifecycle `bootstrap_cases`, 3);
- `cmd/curator/lifecycle_conformance_test.go:306` — `document.UpgradeCases` (manager-lifecycle `upgrade_cases`, 3).
Route all three through `conformancecoverage.Run`/`RunOutcomes`, add pins `umbrella-provider-resolution/cases 14`,
`manager-lifecycle/bootstrap-cases 3`, `manager-lifecycle/upgrade-cases 3` (+ `root-artifacts.tsv` if required). Re-run the
sweep over the WHOLE module (`rg -n --glob '*_test.go' 'range [A-Za-z_.]*\.[A-Z][A-Za-z]*Cases\b|range (entries|family\.' .`
plus your own patterns) and put the full hit list with a disposition per hit in results. One narrowing mutant on a newly
routed family → killed. Bounded runs of `./cmd/curator -run 'Umbrella|Lifecycle'` and the gate scripts. Append "Revision 4"
to results, `task-board resource update` it, then `task-board handoff TASK-260922-3bbvrs --role developer`. A
`run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
