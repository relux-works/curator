# TASK-260825-3n4bjj — Review verdict: ACCEPTED

Reviewer run RUN-260824-6cea80 (claude). Read-only review; no code modified.

## Scope reviewed

Broker and fetch wiring only (sibling deltas in the shared uncommitted tree — `config build-https` CLI (TASK-260825-2gyhq8), `internal/config/buildhttps.go` (168m7o), `internal/gitcred` (1tgpcn), CI gate budget scripts — belong to their own tasks and were not judged here):

- `internal/buildrepo/httpsbroker.go` + `httpsbroker_test.go` — broker, prompt matching, state read, materialization
- `internal/buildrepo/admission.go` — authenticated fetch construction (`acquireNetworkFormat`), `setEnvironmentValue`, per-fetch env scoping
- `internal/install/buildhttps.go` + tests, `internal/install/external.go` — capture/resolution threading and per-repository `GitTool` binding
- `cmd/curator/main.go` — basename dispatch (`main.go:81`), `productionGitTool` askpass source, `productionExternalDeps.BuildHTTPS`

## AC verification

| AC | Verdict | Evidence |
| --- | --- | --- |
| Private HTTPS fetch authenticates end to end against a real repository | PASS | `TestPrivateHTTPSBrokerAuthenticatesRealGitRepository`: real git client, local TLS server requiring Basic Auth, materialized broker answers, locked commit returned only after authenticated requests. Re-run green in this review. |
| Foreign host / prompt / state cases fail closed with tests | PASS | Unit matrix (foreign host, foreign prompt, extra argv, absent secret/state, directory-as-state) plus reviewer probes against the real built binary: symlinked state, relative state path, unknown-field state, no args — all exit 1 with zero output bytes. |
| Anonymous HTTPS byte-identical when no credential is selected | PASS | Authenticated branch gated on `HTTPSCredentials.Selected()`; anonymous argv/env asserted broker-material-free (`TestAnonymousHTTPSArgumentsAndEnvironmentRemainUnchanged`). Under its normal name the binary refuses prompts (usage on stderr, exit 2, no credential bytes). |
| Secret excluded from any diagnostic representation | PASS | `String`/`GoString` redaction on `HTTPSCredentials`, `BuildHTTPSCredentials`, `CapturedBuildHTTPSOverride`, `BuildHTTPSSelection`; asserted through `%v`/`%+v`/`%#v` including nested-in-`GitTool`; state file bytes asserted to carry exactly `host`+`username`; fetch-log secrecy asserted. |
| go test green | PASS | Independent full run: `go test -timeout 30m -count=1 ./...` — 42/42 packages ok, exit 0 (`.temp/TASK-260825-3n4bjj/review-go-test-full-01.log`). Producer log agrees (42/42, exit 0). |

## Definition-of-Done spot checks

- GIT_ASKPASS and core.askPass both point at the wrapper: `TestSelectedHTTPSFetchEnvironmentIsScopedAndOverridesBothAskPassSurfaces` asserts equality of both surfaces on the fetch line and absence of broker material on every non-fetch Git child.
- State/secret ride only in the fetch child env: `fetchEnv` is a copy; init/validation/object-proof children keep the pre-existing env. Wrapper+state live under the per-acquisition `MkdirTemp` root and are removed with it; the secret never touches disk.
- Host binding: credentials bound to `effective.Identity` host; `acquireNetworkFormat` re-checks it against `Source.Identity` and fails closed (`CodeIdentityInvalid`) on mismatch. `effectiveRepository` confirms effective == fetched identity for both declared and substituted network sources.
- Lint/build: `golangci-lint run ./...` 0 issues; `go vet`, `gofmt -l`, `.github/ci/no-broad-suppression.sh` clean (inline `#nosec` suppressions are named-and-reasoned, the sanctioned narrow form); native `go build ./cmd/curator` ok.
- Naming policy (epic): new code/docs reference only this repository and the spec. Clean.

## Reviewed behavior change worth knowing

`productionGitTool` no longer honors an exported `GIT_ASKPASS` as the askpass source; the manager binary is always the broker source. Deliberate and load-bearing: the broker copy must be the manager binary or the fetch environment would hand `CURATOR_BUILD_HTTPS_ASKPASS_SECRET` to a foreign helper, and §12.2 forbids ambient identity-unbound selection. Recorded in LOGBOOK 0316.

## Minor notes (non-blocking)

- `materializeHTTPSCredentialBroker` error attribution collapses to `CodeSourceUnavailable` at the call site even for the `CodeIdentityInvalid` non-absolute-source case; cosmetic.
- `info.Mode()&os.ModeSymlink != 0` in `readHTTPSCredentialState` is redundant after `IsRegular()`; harmless.
- Windows behavior is compile-verified here (cross-build) with unit tests platform-neutral; the git-driven E2E tests skip on Windows. The composite landing task's CI lanes cover it.

## Routing

Accepted → `done`. No `commit_ack` supplied (reviewer archetype); the commit-owning mover for this epic is TASK-260825-1d0eo5 (land-https-credentials-composite), which assembles and commits the scope.
