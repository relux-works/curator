# TASK-260910-5nrmtt revision 6 — ACCEPTED

Candidate tree: dd47598d63b406ea7e51b68447da8d277c123fd1.
Base: 38e62cb8f4b0a427625e17dc6a02cf825f4407a4.
Reviewer run: RUN-260916-044c21. Shell: zsh. Run goal queried before verdict: none; no directives.

## Decision and exact scope

Accept the binding rework-4 change. All 13 candidate paths were byte-compared against Git candidate blobs before and after verification: 13/13 identical. No repository code was edited. Relative to reviewed rev5 tree 2990523057576c695bff9c62f1a3c5dd1965c87d, only four paths change: classifier, transport tests, new reviewer regression, draft documentation. The closed table, Windows refusal, admission, brokers and process control are unchanged.

The framingGlue regexp and pre-split repair are deleted. Classification splits only actual newline boundaries. The positive SSH fixture now terminates diagnostics correctly. TestReviewRev5GluedLineRefuses drives AcquireNetworkResolved with a ready alternate and proves one fetch plus refusal. Prior reviewer probes remain present and passed in the independent run. Full-line table has no open .* or .+ tails; unknown lines poison the entire output; framing alone cannot authorize fallback.

## Independent checks (direct process exit codes)

- go test -p 1 -count=1 ./internal/buildrepo -run 'Review|ResolvedTransport|Classify': exit 1, 303.986 s. Exactly one failure: TestResolvedTransportTotalDeadlineBoundsSlowFetch saw 1 fetch instead of 2. Its elapsed deadline assertion passed. All other selected tests passed, including the new regression, rev1–rev4 probes, 14/14 accepted grammar shapes and 42/42 malformed variants through AcquireNetworkResolved.
- Isolated rerun: go test -p 1 -count=1 ./internal/buildrepo -run '^TestResolvedTransportTotalDeadlineBoundsSlowFetch$': exit 0, 10.949 s. The initial failure is consistent with host contention consuming the six-second shared budget before a second fetch starts, not unauthorized fallback or deadline escape. This is recorded as timing-sensitive fixture behavior, not erased or described as an all-green first run.
- go test -p 1 -count=1 ./internal/gitcred: exit 0, 5.170 s.
- go test -p 1 -count=1 ./internal/buildrepo -run '^Test(SemanticFallbackGateDecidesAllPublishedCases|TransportResolutionPlatformGate)$': exit 0, 0.524 s. These are helper-level bounds, not end-to-end delivery evidence.
- go vet ./internal/buildrepo ./internal/gitcred: exit 0.
- gofmt -l internal/buildrepo internal/gitcred: exit 0, empty output.
- git diff --check: exit 0.

## Independent narrowing mutants: 3/3 killed

Used temporary Go -overlay files outside the repository; candidate bytes never changed.

1. Restore glue repair by replacing 503fatal: with 503 + newline + fatal: before parsing. TestReviewRev5GluedLineRefuses: exit 1, 9.607 s; observed two fetches and nil error. KILLED.
2. Ignore unknown lines instead of returning FailureUnknown. TestReviewRev4ClosedGrammar: exit 1, 17.574 s; DNS plus audit: policy denied made two fetches and succeeded. KILLED.
3. Loosen HTTPS 50x full-line pattern with .* before $. TestReviewRev4ClosedGrammar: exit 1, 14.656 s; 503; object verification failed made two fetches and succeeded. KILLED.

## Acceptance trace and bounds

AcquireNetworkResolved production-entry tests cover strict admission (TestReviewResolvedMustRetainAdmission), forbidden classes (TestResolvedTransportFailClosedClassesStopAfterOneFetch), mixed/truncated/unknown evidence (review probes and TestResolvedTransportTruncatedStderrFailsClosed), two-attempt verified success (TestResolvedTransportFallsBackOnAvailabilityThenVerifies), shared timeout and child pipes (TotalDeadlineBoundsSlowFetch, TestReviewDeadlineIncludesChildPipes), per-attempt real SSH policy binding (TestResolvedTransportSSHAttemptBindsPerAttemptCredentials), alternate locked-content proof (TestResolvedTransportSecondEndpointMustProveLockedContent), legacy behavior (TestResolvedTransportLegacyShapeMatchesLane), clean lane and sanitized errors (TestResolvedTransportKeepsStrictLanePerAttempt, TestResolvedTransportLeaksNothingIntoErrors). These passed, with the timeout test requiring the isolated rerun described above.

Inspected producer results-rev6 and attached rev6-validation.log, not rerun: GitHub run https://github.com/relux-works/curator/actions/runs/35137981615 reports exit 0, Ubuntu/macOS/Windows tests, Ubuntu/macOS race, lint, conformance, naming and gate self-tests successful. Rose-air and candidate-suite jobs skipped. No local Windows runtime or Rose-air verification claimed. Windows resolved-lane refusal was accepted in rev4 and remains unchanged. Full local module suite deliberately not run.

This is the draft internal executor; CLI caller wiring remains a separate leaf and is not claimed delivered here. Frozen schemas, sibling policy checkpoint and revision-2 ports/mirrors/aliases remain untouched. Grammar coverage is the enumerated 14 shapes and 42 malformed variants plus reviewer probes, not a claim to understand every possible Git/SSH diagnostic. No live credentials, runtime-home changes, commits or code edits. Campaign forbids LOGBOOK.md edits: the verdict and board notes persist the timing anomaly instead.

Verdict: ACCEPTED; route through accept_cr revision=6 to integrating, never done from this reviewer.
