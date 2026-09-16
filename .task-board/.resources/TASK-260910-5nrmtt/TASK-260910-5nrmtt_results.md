# TASK-260910-5nrmtt results — rework 1 (rev2)

Producer: developer. Shell for all commands below: bash; exit codes are the
gate's real status (`set -o pipefail` on every piped invocation, otherwise
unpiped with `echo exit=$?`).

## Revision

- Spec: curator-spec main `07e2b41`; the implemented contract
  (`protocol/repository-transport.md` rev 1, `protocol/skillfile-sources.md`,
  `schemas/draft-sources-v1/*`, `conformance/draft-sources-v1/*`,
  `docs/skillfile-sources.md`) is byte-identical to rev1's verified-at
  `871d11b` (empty `git diff 871d11b..07e2b41` over those paths, exit 0).
- Code: Story worktree `task-board/story/STORY-260910-1bhj0g` on top of the
  checkpointed sibling leaf `41901ff` (policy loader/planner; its files are
  untouched — see scope check below). Builds on the landed Skillfile v2
  parser/collections; parser, policy loader, and `internal/gitops`,
  `internal/buildsource` untouched.
- Rework brief: `5nrmtt-rework-1.md`; verdict
  `TASK-260910-5nrmtt_review-verdict-rev1.md`; reproducers
  `TASK-260910-5nrmtt_review-reproducers.patch` applied VERBATIM as
  committed regressions (gofmt-normalized only, not weakened).

## P1-1 — resolved acquisition through strict admission + wrapper binding

- `internal/buildrepo/admission.go`: extracted `admitNetworkRequest` from
  `AcquireNetwork` (trusted Git, canonical source, transport allowlist,
  broker/wrapper presence, SSH selection, lock/tag/ref validation —
  byte-identical diagnostics). `AcquireNetwork` calls it, then its own
  timeout, then `acquireNetworkFormat`: legacy behavior unchanged.
- `internal/buildrepo/transport.go` (`AcquireNetworkResolved`): per-attempt
  order is bind-provider → admit (shared deadline context) → bind SSH
  policy → fetch. Admission/binding refusals record a traffic-free
  `AttemptRecord` classified by code and gate the plan (auth advances or
  exhausts; anything else fails closed with the lane diagnostic).
- `internal/buildrepo/sshbroker.go` (new): `SSHWrapperBase` manager base
  (new `GitTool.SSHBase` field; zero base refuses), `bindSSHWrapperPolicy`
  through the existing `SSHPolicyFor` builder (endpoint shape, resolved
  paths, mandatory pinned host keys), `materializeSSHWrapper` (private
  `curator-build-ssh-wrapper` copy + secret-free JSON state, mirroring the
  HTTPS broker), `IsSSHWrapperInvocation` / `RunSSHWrapper` (exact-tuple
  only, fixed `ExactSSHCommand` argv, exit-code mapping). `TestMain`
  dispatches the wrapper (one line, test-only; the manager-binary dispatch
  line in `cmd/` belongs to the wiring task).
- `acquireNetworkFormat`: when `request.sshPolicy` is set (resolved SSH
  attempts only), the fetch's `GIT_SSH` points at the materialized
  per-attempt wrapper. Direct acquisition never sets it: legacy lane
  byte-identical.
- Reviewer's 4 refusal cases kept verbatim at the production entry point
  (`TestReviewResolvedMustRetainAdmission`, 4/4 pass), plus new
  `TestResolvedTransportSSHWithoutKnownHostsRefusesBeforeTraffic`
  (provider → exhaustion + zero traffic; base creds → lane diagnostic +
  zero traffic) and `TestResolvedTransportSSHAttemptBindsPerAttemptCredentials`
  (real git + real Go wrapper + logging fake ssh: attempt 1 refused →
  availability → attempt 2 host-key → fail closed; ssh argv asserts pinned
  identity/known-hosts/host/path per attempt and cross-attempt isolation).

## P1-2 — positive classification only

- `ClassifyFetchOutput`: new fail-closed integrity row (`mismatch`,
  `integrity`, `corrupt`, `fsck`, `checksum`) and audit row (`audit`,
  `revok`, `canary`, `assurance`, `capabilit`), both checked before
  auth/availability; SSH `permission denied` now requires the daemon's
  parenthesized method list (`sshPermissionDenied` regex) — a bare denial
  is a local failure and stays unknown.
- Truncation is evidence: `boundedWriter` records overflow, `fetchError`
  carries `truncated`, `classifyAttemptError` fails a truncated prefix
  closed as unknown.
- Reviewer's 3 cases kept verbatim (`TestReviewAmbiguousFailureMustNotFallback`,
  3/3 one fetch), plus new `TestResolvedTransportTruncatedStderrFailsClosed`
  (70 KiB availability filler → unknown, 1 fetch; prefix alone proves
  availability) and `TestResolvedTransportAmbiguousFailureMustNotFallbackWhenAlternateReady`
  (same 3 diagnostics against a fully fetchable alternate: any fallback
  would succeed, so 1 fetch + lane diagnostic proves the classifier
  closed the plan). Classifier table extended with 11 fixtures
  (bare/os-error denials, method lists, integrity/audit beats-timeout/
  beats-auth rows).

## P1-3 — total deadline bounds the process graph

- `internal/buildrepo/process_unix.go` / `process_windows.go` (new,
  platform-guarded): own process group per Git child (unix),
  group-kill cancel, direct-kill fallback on Windows.
- `newBoundedGitCommand` (admission.go): `CommandContext` + group-kill
  `Cancel` + 5 s `WaitDelay` pipe-drain bound. Used by `runGitCapture`
  (fetch, init, local lane) and the `ValidateGitTool` version probe.
- Reviewer's fixture kept verbatim (`TestReviewDeadlineIncludesChildPipes`):
  2 s budget, non-exec 5 s sleep now returns in ~2.6 s (was 5.73 s),
  well under the 4 s bound. Existing exec-shape deadline test unchanged
  and green.

## Validation (real exit codes)

- `go test -count=1 -timeout 9m ./internal/buildrepo/` → exit 0
  (`ok`, 100.8 s; full narrow suite incl. all new tests, no skips here).
- `go test -count=1 ./internal/gitcred/` → exit 0 (`ok`, 4.9 s).
- `go test -count=1 ./internal/buildrepo/ -run '^TestReview'` → exit 0
  (all 3 reviewer tests: 4/4 admission, 3/3 ambiguity, deadline ~2.6 s).
- `go vet ./internal/buildrepo/ ./internal/gitcred/` → exit 0.
- `gofmt -l internal/buildrepo internal/gitcred` → no output, exit 0.
- `golangci-lint run ./internal/buildrepo/ ./internal/gitcred/` → exit 0,
  `0 issues.`
- `go build ./...` → exit 0 (whole module still compiles).
- Full module suite deliberately NOT run (host rule: narrow packages only;
  remote gate runs once at handoff).

## Mutant attack (narrowing; bytes restored, proven by sha256)

Production shas recorded before M1 and verified `OK` after every revert;
final tree: transport.go `b0a7a5c3…`, admission.go `6ea48e1a…`,
sshbroker.go `b1577a2a…`.

- M1 drop integrity classifier row: KILLED, exit 1. Classifier
  integrity fixtures fail (mixed case reads availability); the verbatim
  reviewer E2E stays green because its alternate cannot fetch — which is
  why the alternate-ready E2E exists (M1b).
- M1b same mutant vs alternate-ready E2E: KILLED, exit 1. Only
  `integrity-timeout` fails (fallback + success); local/audit pass.
- M2 narrow SSH binding to provider attempts: KILLED, exit 1. Only the
  base-credentials subtest fails (unbound attempt fetches); provider
  subtest still refuses.
- M3 narrow cancel to direct-child kill: KILLED, exit 1. Non-exec child
  shape returns in 6.48 s (bound missed); exec-shape deadline test
  still passes.
- M4 drop truncation check: first run SURVIVED (exit 0) and exposed a
  vacuous fixture — the flood printed to stdout (discarded) instead of
  stderr, and the alternate could not fetch. Fixed both (flood to
  stderr, fetchable alternate): rerun KILLED, exit 1
  ("truncated evidence fell back to a ready alternate and succeeded").
- 5/5 killed, 0 survivors.

## Scope check

Diff touches only `internal/buildrepo` + `internal/gitcred` + tests +
`docs/draft-transport-resolution.md` (updated: binding, classifier,
deadline; the rev1 wrapper deferral is gone). `internal/gitcred`
hunks are rev1's provider namespace, untouched by this rework.
`internal/gitops`, `internal/buildsource`: inspected, intentionally
unchanged (bounded resolution lives on the strict lane; gitops lanes
have no lock/deadline/broker machinery to bind). Sibling checkpoint
files (`internal/config/sourcepolicy.go`, `internal/identity/*`)
untouched. No revision-2 behavior; frozen v1 schemas untouched
(no schema files in the diff).

## Bounds (not covered here)

- Manager-binary SSH dispatch: `RunSSHWrapper`/`IsSSHWrapperInvocation`
  are implemented and tested in `buildrepo` (incl. end-to-end through
  real git via the test-binary copy); the one-line `cmd/curator/main.go`
  dispatch plus populating `GitTool.SSHBase` belong to the wiring task
  (`cmd/` is outside this leaf's scope). Until then a resolved SSH
  attempt with a zero base refuses by design.
- Windows process trees: no group kill; `WaitDelay` still bounds the
  wait while an orphaned helper may linger. POSIX executor tests skip
  Windows (accepted bound, same as rev1); pure gate/classifier/binding
  unit tests run everywhere.
- `newObjectReader` (local `cat-file`, no network, no grandchildren)
  keeps plain `CommandContext`; the bound covers lane Git invocations.
- No live credential export, no runtime-home writes, no network use in
  tests (fixture rewrites + fixture stderr + local fake ssh only).

## Host anomaly (for the record, not a product finding)

Fresh test binaries intermittently stall minutes at launch on this host
(measured: 3m23 s in `dyld` before first test output; one launch stuck
>10 min with zero CPU and was killed). Binaries that launch run
correctly. Mitigation used: longer yields, fixed-path binaries, and a
launch watchdog that retries the same binary. An early M1 observation
(120 s subtest) was this stall, not a product hang: the same mutant
rerun completed in 3.3 s with the predicted narrow failure.

## Checklist note

- Item 8 (logbook): no LOGBOOK.md edit per host rules (control-root and
  LOGBOOK writes are forbidden to producers). Findings recorded here:
  (1) M4 caught a vacuous flood fixture (stdout vs stderr); (2) the
  verbatim reviewer ambiguity E2E cannot see classifier mutants when
  the alternate cannot fetch — the alternate-ready E2E closes that;
  (3) host fresh-binary launch stalls above.

---

# Revision 1 (accepted content preserved below)

# TASK-260910-5nrmtt results — bounded authenticated transport resolution

Producer: developer. Shell for all commands below: bash; exit codes captured
without pipes (`> file; echo exit=$?`), so each is the gate's real status.

## Revisions

- Spec: curator-spec main `871d11bcdfd240a6260d0722503bdd1642a8fce8`
  (verified-at). Transport contract unchanged since `a4fcaf02`:
  `protocol/repository-transport.md` (§1–§3 = revision 1),
  `conformance/draft-sources-v1/semantic-cases.json`
  (`fallback-*`, `pinned-auth`), `docs/skillfile-sources.md`.
- Code: Story worktree `task-board/story/STORY-260910-1bhj0g` on top of
  the checkpointed sibling leaf `41901ff` (TASK-260910-1o9x1f policy
  loader/planner, uncommitted handoff snapshot). Builds on the landed
  Skillfile v2 parser and the sibling's `config.Resolution` shape;
  parser and policy loader untouched.

## What was implemented (revision-1 only; no v2 behaviour)

- `internal/buildrepo/transport.go` (new, 675 lines): `TransportPlan`
  (1–2 closed-grammar endpoints, one identity, opaque provider refs,
  effective fallback; pin already applied by the planner),
  `ValidateTransportPlan` (every violation fails
  `repository_policy_invalid` before any Git call), `FailureClass` +
  `ClassifyFetchOutput` (conservative positive stderr classifier;
  fail-closed rows checked before eligible rows),
  `ClassifyAdmissionCode` (post-fetch lane codes), `AllowSecondAttempt`
  (§2 gate), `AuthProvider` + `CredentialProviders` + `ProviderSet`
  (operator-owned provider table; HTTPS secrets only through the
  trusted broker, SSH paths through the existing validator, explicit
  anonymous), `AcquireNetworkResolved` (one strict-lane fetch per
  endpoint, one shared total deadline, per-attempt provider binding
  that replaces lane credentials, exhaustion as
  `repository_endpoint_unavailable` from a closed vocabulary,
  fail-closed classes returning the lane diagnostic unchanged,
  sanitized machine-private trace records, legacy-shape parity with
  `AcquireNetwork`).
- `internal/buildrepo/admission.go` (+33/-4): fetch stderr captured
  into an internal `fetchError` (bounded 64 KiB, never interpolated);
  `AcquireNetwork` maps it back to the byte-identical legacy
  diagnostic via `laneDiagnostic`. No lane behaviour change.
- `internal/gitcred/provider.go` (new): `curator-provider-https:`
  namespace + `Access.ReadProvider` (strict answer-username match,
  absent degrades to nothing).
- `internal/gitcred/gitcred.go` (+5/-3): `ReadHost` now also excludes
  provider-namespaced answers from the operator's own credential view.
- `docs/draft-transport-resolution.md` (new): boundary note.
- `internal/buildrepo/transport_test.go` (new, 1235 lines),
  `internal/gitcred/provider_test.go` (+76): committed tests below.
- `internal/gitops`, `internal/buildsource`: inspected, intentionally
  unchanged. `gitops.Clone`/`Fetch` (closure/envprofile lanes) have no
  lock verification, deadline, or clean env; retrofitting bounded
  resolution there would rewrite the legacy lanes, and the spec applies
  to existing lanes only when the manager opts in (§1). The bounded
  path is the strict lane (§3/manager-§11 machinery). Locked-content
  verification for network acquisition is the lane's raw-object proof,
  reused per attempt; `buildsource` validates materialized trees.

## Acceptance rows → committed tests (all reach production entry points)

- Existing SSH wrapper / HTTPS broker / lane grammar:
  `TestResolvedTransportKeepsStrictLanePerAttempt` (both fetches carry
  the strict argv, private HOME/config, per-transport allowlist,
  `GIT_SSH` wrapper on SSH; no `insteadOf`),
  `TestResolvedTransportFallsBackOnAvailabilityThenVerifies` (first
  fetch offered the provider broker credential).
- At most two attempts: plan validation
  (`TestValidateTransportPlan`, `TestResolvedTransportRejectsPlanBeforeAnyGitCall`
  with an unrunnable git) + exact fetch counts in every executor test
  (1 or 2, never 0-unless-skipped, never 3; no synthesized URL: the
  loop only indexes the plan).
- Total deadline: `TestResolvedTransportTotalDeadlineBoundsSlowFetch`
  (3s fail + 30s hang inside a 6s budget finishes ~6.3s, under the
  7.5s bound a per-attempt clock would exceed at 9s+),
  `TestResolvedTransportSharedDeadlineStopsSecondAttempt`
  (deterministic resolve-phase expiry; zero traffic).
- Fallback only on positively classified availability/auth:
  `TestSemanticFallbackGateDecidesAllPublishedCases` (all 11
  `fallback-*` cases), `TestClassifyFetchOutput` (33 fixtures incl.
  bare-number, reset, ambiguous-grant, and order adversarials),
  flagship DNS→SSH-success test, `TestResolvedTransportExhaustion*`
  (DNS then auth-rejection).
- TLS, host-key, ref, identity, integrity, audit, unknown, ambiguous
  404 fail closed:
  `TestResolvedTransportFailClosedClassesStopAfterOneFetch` (TLS,
  host-key, 404, repository-not-found, unknown, empty stderr — one
  fetch, lane diagnostic byte-identical),
  `TestResolvedTransportSecondEndpointMustProveLockedContent`
  (moved tag → `build_repository_ref_moved`, no snapshot),
  `TestClassifyAdmissionCode` (ref/integrity/identity/audit code map),
  gate table (integrity/identity/audit/policy tokens stop).
  Audit-denial-after-success cannot re-fetch structurally (success
  returns the snapshot; pipeline owns audit) — no E2E for that row.
- Locked content verified: every attempt runs the full strict lane
  (exact-ref fetch + raw-object proof); flagship asserts commit +
  canonical bytes; moved-tag test proves the alternate cannot
  substitute content; cross-identity plans rejected pre-network.
- Sanitized errors: `TestResolvedTransportLeaksNothingIntoErrors`
  (wrapper echoes the live broker secret into stderr; exhaustion and
  fail-closed errors plus trace dumps never contain it),
  `TestExhaustionErrorIsClosedVocabulary` (no URLs/stderr/newlines),
  `TestAcquireNetworkFetchFailureKeepsLaneDiagnostic` (legacy stderr
  silence for three hostile outputs).
- No ambient unsafe Git config or secret persistence: strict-lane
  per-attempt assertions above; broker state stays secret-free
  (existing shape); anonymous provider offers nothing even with base
  decoy secret (`TestResolvedTransportAnonymousProviderOffersNothing`);
  unconfigured provider records auth with zero traffic
  (`TestResolvedTransportSkipsUnconfiguredProviderWithoutTraffic`,
  `TestResolvedTransportNilProvidersSkipWithoutTraffic`);
  command/path-shaped provider refs rejected pre-network with zero
  Git calls.

## Validation (real exit codes)

- `go test -count=1 ./internal/gitcred/` → exit 0 (`ok`, 3.5s).
- `go test -count=1 -timeout 9m ./internal/buildrepo/` → exit 0
  (`ok`, 35.7s; legacy suites included, no skips on this host).
- `go vet ./internal/buildrepo/ ./internal/gitcred/` → exit 0.
- `gofmt -l internal/buildrepo internal/gitcred` → no output.
- `golangci-lint run ./internal/buildrepo/ ./internal/gitcred/` →
  exit 0, `0 issues.` (First run was exit 1 with 2 revive findings —
  error-last return order, package-comment shape — both fixed in the
  diff, none suppressed.)
- Full module suite deliberately NOT run (host rule: narrow packages
  only; remote gate runs once at handoff).

## Mutant attack (narrowing; bytes restored, proven by sha256)

Baseline checksums recorded before M1; all four production files
verified `OK` after M4 revert. (Two revive fixes landed after the
restore; the full suites above re-ran green afterwards.)

- M1 gate `|| class == FailureUnknown`: KILLED, exit 1.
  Semantic `fallback-unknown` FAILs; E2E `unknown` + `empty` FAIL;
  TLS/host-key/404 subtests still pass (narrow).
- M2 classifier order (availability-timeout before TLS): KILLED,
  exit 1. Only `tls-beats-timeout` FAILs (narrow).
- M3 per-attempt deadline (fresh `WithTimeout` per attempt): KILLED,
  exit 1. Deadline test elapsed 9.6s ≈ 3+6s (predicted signature)
  against the 7.5s bound; resolve-phase deadline test still passes.
- M4 legacy exhaustion (drop the `isLegacyShape` guard): KILLED,
  exit 1. Legacy `failure` subtest FAILs (exhaustion instead of the
  lane diagnostic); `success` still passes (narrow).
- 4/4 killed; 0 survivors.

## Bounds (not covered here, by scope split)

- Wiring: `config.Resolution` → `TransportPlan` conversion and
  routing of policy plans into `AcquireNetworkResolved` belong to a
  later task (config/install out of scope); this leaf has no
  acquisition callers yet. Per-attempt SSH *wrapper-policy*
  materialization (ExpectedHost/RepositoryPath refresh for endpoint 2)
  belongs to that wiring; this leaf binds the per-attempt SSH
  credential selection and reuses the admitted wrapper.
- Fake-git executor tests are POSIX-only (skipped on Windows like the
  existing lane tests); pure gate/classifier/resolver/plan tests run
  everywhere. Cross-platform proof is the remote gate's.
- `runGit` kills only the direct git child; a hung git with lingering
  grandchildren holding the stdio pipes can outlast the context kill
  (pre-existing; the deadline test's wrapper uses `exec`-sleep for
  this reason). Unchanged legacy behaviour, noted not fixed.
- Revision 2 (ports/mirrors/aliases): no v2 behaviour; the plan
  grammar rejects them closed via `ParseSource`.
- No live credential export, no runtime-home writes, no network use in
  tests (fixture rewrites + fixture stderr only).

## Checklist note

- Item 8 (logbook): no LOGBOOK.md edit per host rules (control-root
  and LOGBOOK writes are forbidden to producers). Findings recorded
  here instead: (1) executor initially leaked the internal fetch
  error instead of the lane diagnostic — caught by the fail-closed
  tests, fixed with `laneDiagnostic`; (2) POSIX wrapper must use
  builtins/absolute paths (fetch env has empty PATH) and `exec` for
  killable sleeps (orphaned grandchildren hold Go exec pipes);
  (3) `runGit` grandchild-pipe bound above, pre-existing, untouched.
