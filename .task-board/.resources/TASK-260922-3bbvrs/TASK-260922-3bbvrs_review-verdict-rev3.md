# TASK-260922-3bbvrs — review verdict, revision 3: CHANGES_REQUESTED

Candidate tree 18f9eff6 (base 48da2690), checked in a disposable clone (/tmp/rv3bb).

## Rev2 findings: fixed
- devsub `repository_test.go:66` → `RunOutcomes("skillfile-dev-v2/schema-cases")`, pin 16.
- buildrepo `admission_test.go:194` raw-objects (12), `:244` lfs-pointers (10), `:280` local-config-and-refs (15, bound when no config), `:434` pack-index (8, bound when no index). All four are pinned in conformance-case-counts.tsv.
- Lookup loops that index or pick one named case (godriver/skillspec/buildcache/buildsource/buildmeta/whitelist/skillcheck builddriver files, nodesource P10) are unchanged, which is correct.

## F1 (blocking): the consumer sweep skipped `cmd/`, so three whole-family loops are still unrouted
The producer searched only `internal` (results.md:107: `rg ... internal`). When I repeated the search over `internal` and `cmd`, it found three loops over published rc.12 families. They are not routed through conformancecoverage, have no count pin, and have no reasoned exclusion:
- `cmd/curator/umbrella_conformance_test.go:81`: `for _, tc := range family.Cases` over `vectors/umbrella-provider-resolution.json` `cases` (rc.12 publishes 14).
- `cmd/curator/lifecycle_conformance_test.go:70`: `document.BootstrapCases` from manager-lifecycle `bootstrap_cases` (rc.12 publishes 3).
- `cmd/curator/lifecycle_conformance_test.go:306`: `document.UpgradeCases` from manager-lifecycle `upgrade_cases` (rc.12 publishes 3).
The same manager-lifecycle document already has pinned families (`manager-lifecycle/build-order-cases`, `launcher-cases`), so leaving these out is a gap in the sweep, not a scope decision.

What happens today: if rc.13 drops a bootstrap case, or a case starts failing and someone adds a skip, the gate stays green with no tally mismatch and no ledger row. That is the "vanished case / unlisted failure" class that AC1 and AC3 are meant to close.

Fix: route all three through `conformancecoverage.Run`/`RunOutcomes`, add pins `umbrella-provider-resolution/cases 14`, `manager-lifecycle/bootstrap-cases 3` and `manager-lifecycle/upgrade-cases 3` (plus root-artifacts.tsv if needed), and re-run the sweep over `cmd` as well as `internal`. Suggested command: `rg -n --glob '*_test.go' 'range [A-Za-z_.]*\.[A-Z][A-Za-z]*Cases\b|range (entries|family\.Cases)' internal cmd`.

## Not done in this round
No mutant or test run, because F1 already decides the verdict. The next round should re-apply a narrowing mutant on one newly routed family.
