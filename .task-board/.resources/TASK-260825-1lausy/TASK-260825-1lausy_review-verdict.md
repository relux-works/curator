# Review Verdict — TASK-260825-1lausy per-repository-https-resolution

**Verdict: ACCEPTED**
Reviewer run: RUN-260824-1bc1d7 (reviewer, claude). Third reviewer cycle; the two
prior runs (RUN-260824-50980d, RUN-260824-6ac619) inspected positively but lost
their verdict waiting on the full suite. This run re-inspected independently and
verified first-hand rather than inheriting their conclusion.

## Scope reviewed

Uncommitted epic work in the primary checkout (producers wrote there, per
orchestrator note): `internal/install/buildhttps.go` (new),
`internal/install/buildhttps_test.go` (new), `internal/install/external.go`,
`cmd/curator/main.go:1354` (production capture), consuming
`internal/config/buildhttps.go` and `internal/gitcred` from sibling tasks.

## AC verification (all first-hand unless noted)

1. **Precedence** — `resolveBuildHTTPS` (buildhttps.go:128) resolves per
   effective repository: covering run-wide override → `config.MatchBuildHTTPS`
   longest segment-aware scope (shared generic `longestScope`, buildssh.go:165,
   Spec §6.3) → anonymous. Test
   `TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback` proves override
   beats a matching scope, the longer of two nested scopes wins (secret AND
   username asserted), and an unmatched host stays anonymous.
2. **Host pin (core 12.2)** — buildhttps.go:141 applies the override only when
   `Host == "" || Host == host` (exact equality on the canonical host).
   `TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride` proves a
   non-covered host falls to its configured scope and a non-covered unmatched
   host stays anonymous — both would fail if the pin leaked.
3. **Captured override never renders its secret** — `CapturedBuildHTTPSOverride`,
   `BuildHTTPSCredentials`, and `BuildHTTPSSelection` keep secrets in unexported
   fields with String/GoString redaction; `%v`, `%+v`, `%#v` are asserted
   secret-free AND containing `<redacted>` (so an empty formatter cannot pass
   vacuously) in `TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets`.
   Downstream `buildrepo.HTTPSCredentials` redacts identically and
   `TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch` asserts the whole
   bound GitTool renders no secret. JSON marshal is safe by construction
   (unexported field).
4. **Anonymous is not an error + asymmetry comment** — buildhttps.go:123-127
   states the deliberate asymmetry with SSH (real unauthenticated transport,
   public repositories must keep working). Unmatched rows produce no credential
   and `anonymous` provenance; only a *selected* source with missing material
   errors — absence-of-selection and failure-of-selected-material are distinct,
   as required.
5. **Three fail-closed remedies** —
   `TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies` drives all three
   sources with absent material and fails the run if the gate admits
   (`err == nil` → Fatal): token_env → "set <VAR>"; git-credentials → "'git
   credential approve'" with protocol/host; keyring → "'curator config
   build-https login <scope>'". Each names scope, repository, and exact remedy.
   Parse-time `ValidateBuildHTTPS` guarantees exactly one source per scope, so
   the resolver's default branch cannot see an empty token_env name.
6. **Transport skip** — `needsBuildHTTPS` gates on network-git + https.
   `TestBuildHTTPSResolutionSkipsOtherEffectiveTransports` proves SSH/local rows
   yield zero credentials, zero provenance, and zero credential-store reads even
   with an unpinned override present that would otherwise apply.
7. **Production reachability (no orphan helper)** — capture at process entry:
   `cmd/curator/main.go:1354` mirrors the SSH line above it; resolution inside
   `planExternalBuilds` (external.go:140) with fail-closed propagation;
   credentials carried on `externalPlan` and bound per-fetch via
   `externalGitTool` → `buildrepo.NewHTTPSCredentials`.
   `TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan` drives the real
   `planExternalBuilds` entry point.
8. **Capture immutability** — environment mutated after capture; resolution
   still uses captured values (same test as #3).

## Gates

| Gate | Result | Provenance |
| --- | --- | --- |
| Focused `go test -count=1 -run 'TestBuildHTTPS\|TestCaptureBuildHTTPS\|TestResolvedHTTPS' ./internal/install` | PASS (8/8, exit 0) | rerun first-hand this review |
| `gofmt -l` on both new files | clean | rerun first-hand this review |
| Full `go test ./... -count=1` | 42 packages ok, EXIT=0 | accepted from TASK-260825-1lausy_orchestrator-full-suite.log; verified no Go source file in cmd/ or internal/ is newer than that log, so it covers the exact tree reviewed |
| `go build`, `go vet`, `golangci-lint run` | exit 0 | accepted from implementer evidence (TASK-260825-1lausy_results.md) |

## Architecture fit

Faithful mirror of the accepted SSH resolver shape (capture-at-entry selection
type, per-row resolution, provenance strings, fail-closed configured sources)
while reusing, not reimplementing: `longestScope`, `ValidBuildSSHScope`, and
`internal/gitcred` for material access. The deliberate SSH/HTTPS asymmetry is
documented where the fallback happens. Broker delivery is correctly left to its
separate task; the seam (`externalGitTool` → `buildrepo.HTTPSCredentials`) is
tested from both sides.

## Observations (non-blocking)

- Redaction covers the diagnostic verbs (`%v`, `%+v`, `%#v`, `%s` via Stringer).
  A hostile `%x` on the raw struct would bypass Stringer, but that holds equally
  for the already-landed SSH and buildrepo credential types — a codebase-wide
  convention, not this task's defect, and no call site formats credentials that
  way.
- Provenance messages surface only in dry-run, matching the SSH behavior.

## Handoff

Reviewer-archetype run: no `commit_ack` supplied. Acceptance recorded for the
commit-owning mover; the epic's landing task cuts the composite from origin/main
with only the epic's source and test files, and that mover makes the final
`done` transition with `commit_ack=scope_committed` if this task's transition
requires it.
