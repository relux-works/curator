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
