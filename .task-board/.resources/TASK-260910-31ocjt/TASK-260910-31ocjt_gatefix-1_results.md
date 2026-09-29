# TASK-260910-31ocjt — gate fix 1 (rev1 compile failure)

Context profile: full (role body present).

## Fix
Rev1 (tree 7b224eed) removed `buildrepo.EnvHTTPSBrokerSecret`, but `internal/crossconformance/draftsources_broker_e2e_test.go` still used it, so `go vet ./...` failed on every lane (run 36472221878).
- The e2e test now delivers the secret the production way: an inherited pipe handle named by `CURATOR_BUILD_HTTPS_ASKPASS_HANDLE` (new test helpers `draftsources_broker_handle_{unix,windows}_test.go` mirror `inheritHTTPSBrokerReader`: ExtraFiles fd on unix, `SysProcAttr.AdditionalInheritedHandles` on Windows).
- `runBrokerWithSecret` asserts the child env contains neither the secret value nor `CURATOR_BUILD_HTTPS_ASKPASS_SECRET`, then checks the compiled broker copy answers the password prompt with the secret it got over the pipe.
- New rows: `missing secret handle` (state only, no pipe → silent refusal); `legacy env secret is not answered` (secret only in the removed env var → refusal).
- Rows kept, now using pipe delivery: username/password prompts, plain binary refuses, foreign host/user, bare prompt, argument-count rows, missing/relative/malformed state.

## Commands (zsh, real exit codes)
- `go vet ./...` → 0
- `GOOS=windows go vet ./internal/crossconformance ./internal/buildrepo` → 0 (compile only; Windows runtime is unverified locally)
- `go test ./... -run XXX_NO_TEST` → 0 (every test package compiles)
- `go test ./internal/crossconformance ./internal/buildrepo -run 'Broker|Askpass|HTTPS'` → 0
- `go test -race ./internal/buildrepo` → 0 (214 s)
- `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded` → 0
- Mutant (broker answers from the legacy `CURATOR_BUILD_HTTPS_ASKPASS_SECRET` env var): `go test ./internal/crossconformance -run 'TestDraftSourcesBrokerAskpassDispatch/legacy'` → 1 (killed); after restoring the source, the full dispatch test → 0.

The buildrepo rows from rev1 still pass: TestSelectedHTTPSFetchSecretUsesPipeAndNeverEntersTheProcessEnvironment, TestHTTPSBrokerPipeSurvivesFetchAndAskpassExecOnEveryPlatform, TestPrivateHTTPSBrokerAuthenticatesRealGitRepository.

## CHANGELOG entry (for release prep)
- Security: the HTTPS askpass secret now reaches the fetch child through an inherited pipe handle (`CURATOR_BUILD_HTTPS_ASKPASS_HANDLE`) and no longer through `CURATOR_BUILD_HTTPS_ASKPASS_SECRET`. Grandchildren and environment readers can no longer see it.

No CHANGELOG/LOGBOOK edits. Findings are recorded here (logbook DoD item).
