# TASK-260922-3bbvrs review verdict — revision 4: ACCEPTED

Disposable clone /tmp/rv3bb built from base 48da2690 + CR diff; write-tree == candidate f44fd9f6 (TREE_OK). rev4 patch sha256 70540f44…0156 matches.

## F1 (rev3) — fixed
- cmd/curator/umbrella_conformance_test.go:80 `conformancecoverage.Run(t, "umbrella-provider-resolution/cases", ...)`; lifecycle_conformance_test.go `Run(..."manager-lifecycle/bootstrap-cases"...)` and `Run(..."manager-lifecycle/upgrade-cases"...)`.
- Pins in .github/ci/conformance-case-counts.tsv: bootstrap 3, upgrade 3, umbrella 14; root-artifacts.tsv row for cmd/curator over umbrella-provider-resolution.json + manager-lifecycle.json.

## Own module-wide sweep
`grep -rnE 'range [a-zA-Z.]*(Cases|Vectors|Rows|cases)\b' --include='*_test.go' .` over the whole module (cmd + internal). Every published-vector hit is either routed (dryrun:834, buildcache rejection:264, contextmaterialize:119) or a named-case selector I spot-checked (whitelist builddriver_context:45 filters Name; buildmeta policy:118/150 filter Name/Boundary) or a local table; all match producer dispositions in results.md §"Whole-module sweep". umbrella:98 raw loop only pairs presence bits; tally runs via Run at :80. No unrouted family found.

## Runs (CURATOR_CONFORMANCE_ROOT = curator-spec @ SPEC_PIN dced9b83)
- `go test ./cmd/curator -run 'Umbrella|Lifecycle' -count=1 -v` → ok, no SKIP lines (without the root the lifecycle consumers skip — noted; CI exports root).
- `go test ./internal/conformancecoverage -count=1` → ok.
- Mutant (narrowing, vanished case): `document.UpgradeCases` → `document.UpgradeCases[:2]` → KILLED: `family "manager-lifecycle/upgrade-cases" publishes 2 cases, want pinned count 3`. Restored, clean.

## rev3 → rev4 delta
Applied rev3 patch to base, diffed tree vs rev4 candidate: only conformance-case-counts.tsv (+3), root-artifacts.tsv (+1), results.md, the two cmd/curator test files. Nothing else.

## Residual
- TASK-260922-3bbvrs_results.md is in the candidate tree at repo root; integrator should drop it if it must not land.
