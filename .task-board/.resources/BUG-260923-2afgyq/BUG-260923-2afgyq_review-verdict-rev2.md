# BUG-260923-2afgyq — marker-reader-cross-field-validation

Revision 2 verdict: ACCEPTED. No blocking findings. Delta review of the base refresh.
Base: f40b77c19c01746bda8b9a610358d860f2ad20c5.
Candidate tree: b52821e954825c370ad6c72c69629ded28ed6e1d.
Accepted revision 1 tree: 4eded38359c54177b468b13d7c4934b4c52e4f7b.
Revision 1 base: 2cb29dac8a4c82c5a07d7ca2d107aa6e6e2c93e7.

Substantive acceptance is inherited from attached BUG-260923-2afgyq_review-verdict-rev1.md, which explicitly ACCEPTED revision 1. The generated prompt's reference to a previous rejection is inaccurate. There is no previous rejection mechanism to mark repeat-of.

## Swept surfaces

| Surface | Evidence | Result |
| --- | --- | --- |
| Per-path refresh identity | Two-way multisets of +/- lines, excluding diff headers, compare each revision against its own base. All four paths have empty rev1-only and rev2-only differences. | Held |
| Reader and tests | Both internal/marker files are byte-identical to accepted revision 1. Working files match candidate blobs. Production dispatch and negative proof remain as reviewed in rev1. | Held |
| Ledger | Exact same 20 marker rows removed; five distinct case names, four suite/family groups of five. No additions or unrelated deletions against refreshed base. | Held |
| Trunk hardlink merge | Old-base minus new-base ledger is exactly two rc13 hardlink rows; rev1-candidate minus rev2-candidate is those same two rows. Both stay removed. Other upstream source changes are retained: candidate delta from new base contains only the four advertised paths. | Held |
| CHANGELOG/hygiene | Candidate equals trunk plus exactly the unchanged marker line. Trunk hardlink line remains. No released schema or LOGBOOK candidate changes. git diff --check exits 0. | Held |
| Refresh validation | Three pinned marker/coverage package runs, each exit 0 with -work. | Held |

Findings: none blocking. Free hunt: no additional findings. Repeat-of: not applicable.

## Independently rerun

For each root /tmp/BUG-260923-2afgyq-evidence/{rc13,hashv2,muse}/conformance/v1:
CURATOR_CONFORMANCE_ROOT=<root> go test -work ./internal/marker ./internal/conformancecoverage -count=1 -timeout=3m

Real process exit codes: rc13=0, hashv2=0, muse=0. Raw logs and WORK paths are attached in BUG-260923-2afgyq_review-evidence-rev2.tar.gz. The package suites include direct published cross-field refusal tests and valid-marker fixture tests. Syspolicyd was running, successive crashes=359. Comparison assertions exited 0; exact +/- line counts are in comparison.json.

No mutants, install, CLI, lint, build, full repository, race or cross-platform gates were rerun in this refresh review. Those claims remain bounded by revision 1's attached verdict: independently killed 5/5 per-check mutants, focused install/CLI passing; complete install-package green UNKNOWN after recorded host timeouts. Byte identity justifies reusing that substantive review under the current delta-review brief.

## Exact ledger counting clarification

Old base: 134 data rows; refreshed base: 132; candidate: 112. Total file lines respectively 139/137/117, comprising one header plus four comment/blank lines in addition to data rows. The producer's 138/136/116 counts count every non-header line, including those four comment/blank lines. This is a reporting correction, not a candidate defect. Removed rows are exactly 20, all owned by this bug, representing five distinct named cases; trunk separately removed two rows. No count-pin edits.

The 108 worktree board-path diagnostics are outside the candidate delta; no board-path changes are accepted as repository content. Reviewer changed no source or board files directly; evidence is attached through the CLI. Run goal queried: not goal-bound. Acceptance uses accept_cr revision=2, routes to integrating, and supplies no commit_ack or done transition.
