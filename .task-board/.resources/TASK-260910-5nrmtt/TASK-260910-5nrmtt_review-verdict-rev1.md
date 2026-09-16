# CHANGES_REQUESTED — TASK-260910-5nrmtt CR revision 1

Candidate: `3b7d4c1b251d4bf8c08e215514374f3ed168a8c4`; base: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`. Independent reviewer; no product or committed-test files modified. Verdict: **changes_requested**, route **to-dev**. No external/human decision is needed.

## Findings requiring rework

1. **P1 — resolved acquisition bypasses strict admission** (`internal/buildrepo/transport.go:536`). Calling `acquireNetworkFormat` directly skips every admission check at `AcquireNetwork` lines 247–282: trusted Git validation, canonical source, broker/wrapper presence, SSH credential selection, locked-commit validation and exact/substitution ref validation. The private helper does not repeat them. `TestReviewResolvedMustRetainAdmission` proves 4/4 refusals are bypassed: missing SSH wrapper, missing SSH credentials, missing HTTPS broker, invalid substitution ref kind. Each legacy control refuses, while resolved acquisition fetches once and returns a valid snapshot. Preserve/refactor all admission checks into the resolved production path, with the shared timeout intact, and add these regressions. Also bind the selected SSH credentials into the actual wrapper policy for each attempt: assigning `GitTool.SSHCredentials` alone never reaches `cleanGitEnvironment` or a wrapper-policy builder. The existing stand-in rewrites SSH to file:// and sets the wrapper to /usr/bin/false, so success there proves no actual SSH authentication.

2. **P1 — substring classification permits forbidden fallback** (`internal/buildrepo/transport.go:190–203`). Generic `permission denied` is not positive SSH authentication rejection; generic `timed out` can describe audit, integrity or partial-transfer failures. `TestReviewAmbiguousFailureMustNotFallback` drives the executor with (a) `fatal: cannot create temporary file: Permission denied`, (b) `remote: audit denied: operation timed out`, (c) `error: object hash mismatch` followed by `fatal: connection timed out`. All 3/3 cases perform two fetches and return success. They must stop after one. Positively identify transport/authentication diagnostics and reject mixed, ambiguous and integrity/audit evidence. Bounded stderr capture also discards truncation information; do not authorize fallback from an incomplete diagnostic prefix.

3. **P1 — total deadline does not bound a normal child-process tree** (`internal/buildrepo/admission.go:476–484`, resolved call at transport.go:536). `exec.CommandContext` kills only Git's direct process; an SSH/helper grandchild retaining stderr keeps `Run` waiting. `TestReviewDeadlineIncludesChildPipes`, using the existing fake transport with a non-exec 5-second sleep and a 2-second budget, returns after **5.730s** (test exit 1). This is a pre-existing runner limitation, explicitly acknowledged by the producer, but the new feature cannot claim the required total bound while inheriting it. Bound cancellation/pipe draining for the whole admitted process graph and test the non-exec child shape. The producer's exec-sleep fixture avoids this failure shape. An initial 500ms probe passed because it did not reliably reach the sleeping fetch; it is not evidence of the deadline property. The larger-budget reproducer reaches and exposes the failure.

## Independent validation and evidence identity

Commands ran in zsh, with actual process exit codes observed. No full-module suite and no repeat hosted gate.

- `go test -count=1 ./internal/buildrepo`: exit 0, 40.289s.
- `go test -count=1 ./internal/gitcred`: exit 0, 4.879s.
- `go test -count=1 ./internal/config ./internal/identity`: exit 0, 2.734s / 1.632s.
- `go vet ./internal/buildrepo ./internal/gitcred ./internal/config ./internal/identity`: no diagnostics, combined command exit 0.
- `gofmt -l internal/buildrepo internal/gitcred internal/config/sourcepolicy.go internal/config/sourcepolicy_test.go internal/identity/draft_sources.go internal/identity/draft_sources_test.go`: no output, exit 0.
- Reviewer attack overlay: `go test -count=1 -overlay .temp/review-TASK-260910-5nrmtt/attack.json ./internal/buildrepo -run '^TestReview'`: exit 1 (4 admission and 3 classifier failures). Deadline test subsequently added and run separately with `-run '^TestReviewDeadline'`: exit 1.
- All candidate ordinary files and symlink bytes verified against the tree. Gitlink `agents/skills/skill-go-testing-tools` is uninitialized here and was excluded from byte equality; no dependency/platform claim inferred from it. No tracked candidate bytes changed by review or mutants.
- Hosted validation accepted as attached evidence only: run 35086804037, gate commit `70469c7b321f63154d480cb9283c68564d2c4ef1`, whose tree independently resolves to the exact candidate. Attached log reports Linux/macOS/Windows tests, race where configured, lint and other gates successful, exit 0; rose-air skipped. POSIX transport executor tests skip Windows, so Windows execution coverage for them remains unverified despite green hosted jobs.

## Narrowing mutants

Used Go `-overlay` files, leaving original bytes untouched (no restore necessary). Both ran uncached; logs attached.

- M1: narrow fail-closed gate by permitting only FailureUnknown additionally. `TestSemanticFallbackGateDecidesAllPublishedCases/fallback-unknown` and executor unknown/empty cases fail, exit 1: **killed**.
- M2: narrow host-key recognition by removing only the `host key` token. Classifier host-key and host-key-beats-auth cases fail; executor host-key class assertion also fails, exit 1: **killed**. The plain-host-key executor case still stops as unknown; the helper mixed-signal case is what proves fallback exposure under this mutant.
- Mutation result **2/2 killed**, 0 survivors. This does not establish complete contract coverage: the reviewer probes above expose **7/7** admission/classification holes and **1/1** deadline hole outside those existing cases.

## Scope and acceptance assessment

The full 18-path CR includes the checkpointed policy sibling (`41901ff`). The actual delta above that checkpoint is seven files in buildrepo/gitcred plus docs; it stays within this leaf's scope. Frozen schemas are unchanged; docs explicitly label draft/opt-in. No revision-2 behavior introduced.

Existing test evidence supports plan count/identity rejection (`TestResolvedTransportRejectsPlanBeforeAnyGitCall`), DNS fallback with raw-object proof (`TestResolvedTransportFallsBackOnAvailabilityThenVerifies`), moved-tag refusal (`TestResolvedTransportSecondEndpointMustProveLockedContent`), error redaction (`TestResolvedTransportLeaksNothingIntoErrors`), exhaustion and anonymous-provider behavior. It does not establish the full AC: strict admission, positive-only classification and real total deadlines fail as above. Audit/integrity/identity fallback rows are partly helper-direct mappings, not all end-to-end refusal proofs. `AcquireNetworkResolved` currently has no non-test callers; docs explicitly defer caller wiring. This review does not require unrelated legacy-lane redesign or release activation, but the executor itself must enforce its advertised boundary before acceptance.

Rework should retain passing positive/negative/legacy cases, add the attached production-entry regressions, exercise actual wrapper-policy/provider binding, then hand off a new revision for independent review. Do not solve these failures by weakening assertions or by using a fake process shape that cannot exhibit them.

Findings recorded on the board as the logbook-equivalent permitted by campaign rules; LOGBOOK.md edits are expressly prohibited. Run goal queried before verdict: not goal-bound.
