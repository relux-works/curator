# BUG-260923-2afgyq — marker-reader-cross-field-validation

Revision 1 review verdict: ACCEPTED. No blocking finding in the reviewed scope.
Base: 2cb29dac8a4c82c5a07d7ca2d107aa6e6e2c93e7.
Candidate: 4eded38359c54177b468b13d7c4934b4c52e4f7b.
The four candidate file digests matched producer evidence before and after review.
Reviewer made no repository changes; mutants used Go overlay files in /tmp.
Run goal queried before verdict: this run is not goal-bound.

## Swept surfaces

| Surface | Result |
| --- | --- |
| Production reader dispatch | Read -> ReadState -> validMarker -> validBuildState -> validV3Build reaches the checks for v3/v4; core-v5 behavior is preserved. |
| Unsubstituted external build | Identity kind/value, object format and commit bind to the declaration; nil safety follows validDriverBuild. |
| Local/network substitutions | Typed identity kinds and ref shapes are enforced. Local builds retain their existing path. |
| SHA-1/SHA-256 revision | Width and hexadecimal checks bind the structured revision to its object format. |
| Compatibility | All 100 published valid fixtures pass across three pinned suites and v1 through published v5. Package-v5 path is unchanged. |
| Tests | Foreign-writer bytes exercise Read directly; ReadState present/invalid behavior is also covered. |
| Ledger/hygiene | Only this bug's rows removed; one Unreleased CHANGELOG line; no schema or LOGBOOK edits. |

## Independently executed evidence

All commands preserved process exit codes. No producer test result substitutes for the checks below.
Fixture roots: /tmp/BUG-260923-2afgyq-evidence/{rc13,hashv2,muse}/conformance/v1.
Root manifest digests were verified against source-identity.json.

- With each root: go test ./internal/marker ./internal/conformancecoverage -count=1 (rc13/muse also -v): exit 0 each.
- Direct authoritative fixture details: 26/26 valid rc13, 37/37 hashv2, 37/37 muse; 10/10 invalid v3/v4 regressions per root, 30/30 total. This includes all five named v4 cases in every root.
- Targeted cmd/curator classification, marker, status and repair tests: exit 0; cmd.log.
- Targeted internal/install tests with rc13, -count=1 -timeout=4m -v: exit 0 (55.257s); install-focused-rc13.log. Seven top-level tests cover legacy marker-v3 installation, external command collision, global mixed build order, authoritative mixed build cases (6/6), revision substitution object format, marker generation read states, and moved-tag invalid-marker behavior.
- golangci-lint run ./internal/marker/... ./internal/conformancecoverage/...: exit 0, 0 issues.
- go vet ./internal/marker ./internal/conformancecoverage ./cmd/curator: exit 0.
- go build -o /tmp/BUG-260923-2afgyq-review/curator ./cmd/curator: exit 0.
- git diff --check: exit 0.

## Negative proof

I independently killed 5/5 per-check mutants through the matching published v4 fixture: declared commit binding, local identity kind, network identity kind, SHA-1 width, SHA-256 width. Each go test exited 1 specifically with “Read admitted the published invalid marker”. Width mutants relax only their own object-format arm; the binding mutant retains identity equality while dropping commit equality. This proves narrower failures as well as individual checks. Exact overlay replacements, commands, logs and exit codes are in the review evidence archive. No mutation touched the worktree.

## Exact ledger accounting

“Five rows” describes five distinct case names, not this candidate's physical ledger shape. Exactly 20 physical rows were removed: 5 hashv2 v4, 5 rc13 v3, 5 rc13 v4, 5 muse v4. All 20 are owned by this bug and every corresponding exact published case passed. There are zero additions or unrelated removals, and no count-pin edits. Retaining the duplicate rows after these passes would leave stale gaps. This is consistent with the task's cross-version validation and ledger intent; it is not a claim that the diff contains only five deleted lines.

## Anomalies and bounds

The initial combined marker/coverage/install command exited 1: install timed out at 600 seconds during a host execution stall. Even fresh shells and launchctl were delayed; after recovery syspolicyd reported running and successive crashes increased from 353 to 354. The timeout stack was in draft-audit fixture command execution.

Two broad install subsets of 70 top-level tests each subsequently exited 1 at their 4-minute deadlines: subset 0 was in script-audit staging/journal work; subset 1 was in draft-audit tests. Remaining broad subsets were not run. These timeouts are retained as failed attempts, not passes; a complete install-package green is UNKNOWN. No behavioral assertion failure attributable to the changed reader was observed in these attempts. Acceptance is limited to the passing marker suites and explicitly scoped install/CLI consumer tests.

The first focused install attempt used hashv2 and exited 1 because that root has no committed count pin for external-repository-lifecycle/mixed_build_cases (six other selected tests passed). Inspection showed the pin belongs to rc13. Rerunning the same seven tests against rc13 passed, including all six published mixed-build cases. No pin, test or source was changed to achieve this result.

Producer evidence was read for context, including its 3 additional identity-binding mutants; those extra mutants were not independently rerun and are not included in my 5/5 claim. Full repository, cross-platform and race validation were not rerun.

Acceptance routes to integrating via accept_cr; no done transition or commit acknowledgement is supplied. The conditional changes-requested checklist item is not applicable to this accepted verdict.
