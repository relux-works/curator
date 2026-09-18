# TASK-260917-16l2md — revision 5 review verdict

Verdict: **accepted**, revision 5. Route with accept_cr to integrating; acceptance is not landing.

## Identity and cleanup

Reviewed base `3c45d4bbed81348dfc814bac98b87a58aadb02af`, candidate tree `559e88e475d813ebe6c28e3dcaa798af028c0c74`. Downloaded patch SHA256 is `18a5cd686f64c8e4f17ab74240b703db86a4a24e4de910ec110bf375b3cc489d`, matching the assignment. Independently compared revision-4 tree `f7335cbce8e7e24f8006fc069a9993a08cb4cbaa`: the entire repository delta is deletion of root `curator`, mode 100755, blob `39e528eb7c8ca06ed87d9c263cd3ed3469b540b7`. Every remaining blob and mode is identical (stronger than per-file patch-id). Results §12 acknowledges the cleanup and corrects §11.6's count; this board resource is outside the repository tree.

All 40 changed worktree paths match candidate blobs byte-for-byte. Status lists only source/test/docs/CI changes, including three legitimate untracked additions; no root executable, .review, LOGBOOK, coverage or other untracked output. Candidate was never modified. Tests ran in a disposable archive under /tmp, using exact spec `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

## Per-item review

| Item | Assessment |
|---|---|
| Sole revision-4 rejection | Closed: captured executable removed, no other repository changes. |
| Union fidelity | Retained byte-for-byte from the independent revision-4 review, including candidate overlap resolutions, parser/lock maps, status rows and vector drivers. Prior substantive review and its input-patch comparison are reused, not represented as a new full semantic audit. |
| Pin/scope | One rc.12 SPEC_PIN, five env.SPEC_PIN suite checkouts; release.yml and ledger unchanged. No added spec bytes, skip filters, E1/R1/S6 scope or build artifacts. |
| E2 | Unchanged default drop/error, direct/overlay/waiver admission, error-only lock direction, null rejection, posture and exact comparisons. Schema/admission drivers replayed. |
| E4 | Unchanged revision A warning and B refusal, managed-directory refusal, trust-root identity and posture. Null Load probe replayed with omitted/empty controls. Local umbrella identity tests replayed; Windows test read, Windows execution accepted from rev5 hosted evidence. Pinned §11 explicitly keeps PATH selection in A; the brief's blanket “never PATH” is not the normative rollout rule. |
| S4 ordering | Exact predecessor publication probe passes. Committed install/reinstall/update order, post-emission failure and non-fatal sink tests pass. CLI sinks and all five pre-publication carrySurfacing sites unchanged; sink emission clears returned rows to prevent double printing. |
| S4 passthrough | Both profiles and production Resolve bounds replayed; s4-warn remains shipped. Null remains explicit unbounded; absent behavior follows rollout §10.3. |
| Docs/CHANGELOG | Operator guide and E2/E4/S4/pin notes unchanged from accepted substantive review. Cleanup acknowledged in results. |
| Coverage boundary | Full shell package passes; prior review explicitly identifies no shell-hook-trust vector driver in this candidate. That S6 family remains outside this task; package success is not claimed as its vector execution. |

## Independent replay and accepted evidence

Bash with set -o pipefail; all Go tests -count=1, rc.12 conformance root. Commands and transcripts are in TASK-260917-16l2md_review-evidence-rev5.zip.

- go build ./..., go vet ./..., gofmt -l internal cmd: each exit 0; formatting output empty.
- Full config, contextresolve, contextmaterialize, contextaudit, shell, interop/environments and envfragment packages: exit 0.
- Targeted envprofile vectors and emission-order tests: exit 0.
- Targeted CLI umbrella, status and surfacing tests: exit 0.
- Exact predecessor TestReviewerSurfacingBeforePublication and TestReviewerProviderDirectoriesNullRejected probes: each exit 0.
- golangci-lint for config/envprofile/CLI: exit 0, 0 issues.
- Gate self-test: exit 0; targeted config/schema/admission/null vector replay: exit 0. Profiles and CLI verbose logs have respectively 31 and 39 PASS entries, zero skips; vector log also has zero skips.

Full envprofile and CLI suites, Linux/Windows/race are not independently replayed in this cleanup review. Accepted attached revision-5 hosted validation, run 35299608295, all required lanes success, exit 0; optional candidate/rose-air lanes skipped. Prior revision-4 substantive review is TASK-260917-16l2md_review-verdict-rev4.md with its evidence bundle. No full remote gate rerun.

## Narrowing mutants

Replayed three predecessor overlays on the byte-identical production files; **3/3 killed**, each exit 1:

| Mutation | Committed test detecting it |
|---|---|
| E2 error refuses only when len(dropped)>1 | TestSystemPromptErrorRefusesFirst and system-module-transitive-error via SystemPrompt |
| E4 managed-directory refusal only in revision B | Three umbrella vectors, TestUmbrellaRefusedDirectoriesBothRevisions, symlink/case identity tests |
| S4 unlisted names admitted when effective list empty | Enforce absent/empty tests, passthrough vector, production Resolve TestResolvePassthroughKnobs/empty-list-bounds-all |

Bound: three targeted predicates, not exhaustive mutation coverage. The two reviewer probes and committed order tests are replayed separately against unmutated production code.

## Lifecycle

Run goal queried: no goal binding. Directives: none. No code edits, commits, pushes, branch changes or commit_ack. Campaign forbids LOGBOOK.md edits; this outcome and board notes preserve the review. Evidence attached before accept_cr.
