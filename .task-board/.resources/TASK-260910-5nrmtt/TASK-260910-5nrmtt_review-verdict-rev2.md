# CHANGES_REQUESTED — CR-TASK-260910-5nrmtt-2 revision 2

Reviewer run: RUN-260916-1ba737. Route: to-dev. No acceptance or commit acknowledgement.

## Exact revision and scope

Base f38ee3946110eeeee6bad22831118da117ad6292; candidate tree c65570b0c33bcd5233be105c1ecaf2b183226b10. Independently hash-compared all 23 CR paths against working files: 23/23 equal. Hosted gate commit a64e07d1b27b5804dd66d1e91340e9467d68e9e1 resolves to that exact tree. The 11 additional policy/identity paths are the sibling checkpoint; the current uncommitted transport change has 12 paths. No code or candidate bytes modified during review. Go overlays under /tmp held probes and mutants.

Contract: repository-transport.md revision 1, sections 2–3, read from the curator-spec checkout; Skillfile sources opt-in scope retained. No frozen schema or revision 2 behavior introduced. Read producer results and rev2 hosted validation log.

## P1: mixed/unknown diagnostics still authorize fallback

internal/buildrepo/transport.go:185–217, particularly broad substring checks at 203–216. Adding integrity/audit keyword deny lists did not implement positive diagnostic identification. Any remaining unknown text is ignored if an eligible substring exists anywhere, including a local filename.

Independent TestReviewRev2MixedUnknownMustNotFallback drives AcquireNetworkResolved with a fully fetchable alternate. All 3/3 cases fail: two fetches and nil error (successful acquisition), where the contract requires one fetch and refusal:

1. Valid DNS error followed by `fatal: unexpected protocol response` (mixed/unclassified evidence).
2. `fatal: unable to read object 0123456789012345678901234567890123456789` followed by `fatal: connection timed out` (object failure plus timeout).
3. `fatal: cannot create temporary file 'connection timed out': Permission denied` (local filesystem failure, not endpoint availability).

Reproducer patch and exact output attached separately. Fix diagnostic parsing to recognize complete, context-appropriate transport/authentication errors and reject unrecognized/mixed/malformed evidence. Merely adding these three strings to a deny list would retain the defect class. Preserve these production-entry probes and add grammar/unknown-tail coverage. The original three ambiguity fixtures pass but cover only their selected keywords.

## P2: Windows cancellation still does not bound the process graph

internal/buildrepo/process_windows.go:12–20 kills only the Git process. admission.go sets an additional five-second pipe WaitDelay; this neither kills surviving SSH/helper descendants nor enforces a two-second total deadline. The producer explicitly reports that orphaned helpers may linger and POSIX executor tests skip Windows. This is an implementation limitation, not an accepted exemption in the task contract or rework brief. Provide Windows process-tree lifetime control (or refuse this resolved lane there before process creation until supported) and a Windows-native descendant fixture. This finding is code-inspected, not runtime-reproduced on Windows; Windows behavior remains unverified here. Hosted Windows green does not exercise the skipped process-tree tests.

## Independent validation

Shell zsh. Commands unpiped; statuses collected directly.

- `go test -count=1 -p 1 -timeout 4m ./internal/buildrepo -run 'TestReview|TestResolvedTransportAmbiguous|TestResolvedTransportTruncated|TestResolvedTransportSSH'`: exit 0, 9.365 s. Original 4/4 admission and 3/3 ambiguity cases pass; POSIX descendant deadline and SSH binding checks pass.
- `go test -count=1 -p 1 -timeout 4m ./internal/buildrepo`: exit 0, 50.314 s.
- `go test -count=1 -p 1 -timeout 2m ./internal/gitcred`: exit 0, 3.382 s.
- `go vet ./internal/buildrepo ./internal/gitcred`: exit 0.
- `gofmt -l internal/buildrepo internal/gitcred`: exit 0, no output.
- `git diff --check`: exit 0.
- New reviewer probe via `go test -overlay /tmp/TASK-260910-5nrmtt-review2/repro.json -count=1 -p 1 -timeout 2m ./internal/buildrepo -run '^TestReviewRev2MixedUnknown'`: exit 1, 3/3 forbidden fallbacks reproduced.

Hosted evidence accepted as already attached, not rerun: rev2 validation log reports ubuntu/macos test+race, Windows test, lint, interop and gate self-tests successful (run 35098788892). rose-air skipped; no ARM64 or Windows process-tree runtime claim. No full local module suite run.

## Narrowing attacks

All compiled with Go overlays; originals remained byte-identical, so no restoration writes were needed.

- M1b: narrow audit refusal to revocation/canary/assurance/capability (drop literal audit). Alternate-ready production-entry test fails on audit-timeout with successful forbidden fallback. KILLED, exit 1.
- M2: narrow truncated-evidence refusal to prefixes containing certificate. TestResolvedTransportTruncatedStderrFailsClosed fails with successful forbidden fallback. KILLED, exit 1.
- Additional M1: drop canary recognition only. Classifier test fails on audit-canary. KILLED, exit 1; helper-level diagnostic coverage only, not a fallback proof.

2/2 production-entry narrowing mutants killed; 1/1 additional helper diagnostic mutant killed. New independent ambiguity probes: 0/3 refusals, 3/3 unwanted successful fallbacks. Existing green fixtures are not evidence of arbitrary mixed-output refusal.

## Acceptance trace and bounds

Strict admission/brokers/grammar: TestReviewResolvedMustRetainAdmission, TestResolvedTransportKeepsStrictLanePerAttempt, SSH binding and known-host tests. Attempt count/deadline: plan rejection tests, executor fetch counts, TestResolvedTransportTotalDeadlineBoundsSlowFetch and TestReviewDeadlineIncludesChildPipes (POSIX only). Lock verification: TestResolvedTransportSecondEndpointMustProveLockedContent and availability-success fixture. Sanitization/ambient credentials: TestResolvedTransportLeaksNothingIntoErrors, TestResolvedTransportAnonymousProviderOffersNothing and strict-lane per-attempt checks. Legacy regression: full narrow buildrepo suite and TestAcquireNetworkFetchFailureKeepsLaneDiagnostic.

Fallback acceptance is NOT met (P1); cross-platform process lifetime is NOT established (P2). Resolver and SSH manager dispatch have no cmd caller yet; producer assigns that wiring to a separate task. No claim that the CLI feature is delivered by this leaf. Gitops/buildsource unchanged. No real credential export or runtime-home changes. Findings recorded in this task-scoped verdict and board notes; LOGBOOK.md left untouched per campaign rule.

Goal queried before verdict: run is not goal-bound. Rework is recoverable; no human decision or external blocker required. Preserve sibling policy checkpoint, implement fixes, rerun narrow checks and publish a new candidate for independent review.
