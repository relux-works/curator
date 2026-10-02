# Accepted — spec-owner-review-triage-101-106-112

Task: TASK-261002-1zk11s. CR revision 2; candidate f9f349fb62bff556dd93f4f798e1b76fe00bdd7c; base 2cb29dac8a4c82c5a07d7ca2d107aa6e6e2c93e7. Reviewer verdict: accepted. No blocking findings.

## Swept surfaces

| Surface | Evidence and finding |
|---|---|
| #101 | Curator internal/godriver/session.go:42 and :256 permits only 1.25 and refuses other parsed families. Spec protocol/core.md:670–698 requires tested families and fixed probes. Recommendation correctly describes a normative opt-in change. |
| #106 | session.go:49–52 and :483–533 establishes no downloads and explicit CURATOR_GO/GOROOT/runtime-root selection. Card I1 gives three scope/default choices, retention and replay costs, and explicitly leaves provisioning unimplemented pending Ivan. |
| #107 | Spec protocol/core.md:1540–1548 excludes user PATH and wrappers; :686–698 fixes three probes. Curator session.go:536–556 rejects non-native wrappers. Card I3 correctly treats discovery as a trust change; real shim compatibility remains unknown. |
| #108 | internal/skillspec/parse.go:520–528 excludes dependencies.tools. Card I2 distinguishes publisher responsibility, signed registry records and upstream evidence; choices are real, and implementation explicitly waits for Ivan. |
| #109 | internal/buildrepo/admission.go:361–368 selects exact tag/lock; :433–437 compares terminal commit. Remote reachability is correctly distinct from successful object fetch; proposed remote diagnostic remains later and bounded. |
| #110 | Spec protocol/skillfile-sources.md:83–96 states physical output exclusion. Curator closure/resolve.go:113 and :908–915 builds scope-relative roots. staging/boundaries.go:107–177 implements canonical/SameFile containment and name-based .git detection. snapshot/boundaries.go:153 and :174 invoke it in real admission. No gitfile-target parser/registration was found in this inspected path. This is a bounded absence finding, not a claim about every Git helper in the repository. |
| #111 | Spec profiles/manager.md:392–415 contains the narrow assembly/directive allowances and replacement restriction; triage preserves that boundary. |
| #112 | Spec protocol/core.md:1279–1305 binds acquisition to terminal commits and bounded peeling. admission.go:781–822 proves and peels before comparison. The report correctly avoids claiming that acquisition establishes first-write coverage. |
| Deliverable | 8 issue rows; 3 owner cards; 10 ordered leaf rows, with L8 explicitly split per issue. Kind, state, recommendation, effort, release and consumer/lockstep impact provided for each. |
| Change scope | Exact candidate contains only .research/261002_spec-owner-review-triage-101-106-112.md. LOGBOOK.md is byte-identical to base. Working report matches candidate bytes. No code edits by reviewer. |

Spec citations checked against e41c561b3300a0a4d3425437fddc5a7048d7b11e; curator citations checked against the CR base, using git show with checked subprocess exit codes (all 0). More than five citation spans were verified across both repositories. Source links are the immutable links in the reviewed report. Re-read all eight GitHub issue bodies and returned comments through gh issue view --comments --json; each exit 0. Only #107 returned a comment; card recommendations address the relevant owner questions. No issue mutations or external messages.

## Fresh reviewer verification

- `go test ./internal/godriver -run '^(TestSelectionPriorityAndNoPathLookup|TestUnsafeToolchainCandidatesFailBeforeExecution|TestProbeFailuresCleanPrivateStateAndHaveStableDiagnostics|TestAuditedVendorAllowancesAreWithheldFromAReplacedModule)$' -count=1`: exit 0.
- `go test ./internal/buildrepo -run '^TestTaggedAcquisitionUsesOnlyExactTagAndChecksTerminalCommit$' -count=1 -v`: exit 0; SHA-1 and SHA-256 cases passed.
- `go test ./internal/staging -run '^(TestWithinCatchesSymlinkAliases|TestIsOutputPathPrunesGitAndOutputs)$' -count=1 -v`: exit 0, both selected tests passed.
- Coverage: 7/7 selected top-level tests, with refusal cases included. These rerun the report's bounded baseline checks; no full suite, real future compiler/shim, provisioning, csk runtime or Windows matrix attested.
- `git diff --check BASE CANDIDATE`: exit 0. `git diff --exit-code BASE CANDIDATE -- LOGBOOK.md`: exit 0. Python byte/path assertions: exit 0.
- Fresh origin main advertisement and exact-main fetch both resolved f40b77c19c01746bda8b9a610358d860f2ad20c5, exit 0. Upstream paths do not overlap the research document or the cited implementation files. The report identifies its immutable earlier baseline honestly; integration freshness remains producer-owned.
- `task-board spawn goal` reports this run is not goal-bound. Directive check reports none.

## Nonblocking historical note

The unchanged research report mentions its former LOGBOOK entry and validation of that entry. Revision 2 removes that entry entirely. Treat those phrases as historical producer activity, not the final delta. The binding rework instruction explicitly required preserving this research report unchanged; this note clarifies the final state without requesting contrary rework. This task-scoped outcome also records the review anomaly in place of editing repository LOGBOOK.md.

A missing reviewer-role reference and two exploratory board schema/projection errors were not accepted as evidence; corrected task/checklist reads succeeded. No failed read was used to infer absence.

Acceptance is for the research triage, not permission to implement owner-gated features. Conditional rejection-routing checklist item is N/A because the verdict is accepted. Route through accept_cr to integrating; do not mark done or supply commit_ack.
