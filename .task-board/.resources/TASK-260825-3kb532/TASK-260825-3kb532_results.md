# TASK-260825-3kb532 — install precheck and candidates

## Delivered behavior

- Collects every unmatched effective HTTPS build repository and runs the
  credential resolver before any `Acquire` call.
- On an operator terminal, presents presence-only discovery of the operator's
  existing Git HTTPS credential plus an explicit "enter a token now" option.
  Discovery cannot select or return a secret; material is read only after the
  operator selects it.
- Uses one shared persistence question for HTTPS and SSH. Enter accepts the
  narrowest repository-namespace scope, `r` means this run only, and `q`
  aborts. A chosen scope must actually cover the repository identity.
- A this-run-only answer returns run-local material but never calls the config
  or credential-store persistence callback. This is covered for both HTTPS
  credential surfaces and the latent SSH path.
- A saved entered token is stored through the manager-namespaced Git credential
  entry and then selected in `build_https`; a saved existing credential records
  only the `git-credentials` source and never copies the operator secret.
- Off a terminal (and on dry run), HTTPS keeps a nil resolver and unmatched
  repositories continue anonymously.
- Updated the SSH prompt documentation and README credential summary.

## Negative and production-path evidence

- `TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface`
  covers existing-host and entered-token choices.
- `TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore` compares the
  real config and an isolated real Git credential-store file byte-for-byte.
- `TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting` and
  `TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig` cover the SSH regression.
- `TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential` proves a
  resolver that makes no selection leaves the repository anonymous and never
  reads host credential material.
- `TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch` drives
  `planExternalBuilds` and proves abort prevents the first acquisition.

## Validation (standalone commands and real exit codes)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/install` | 0 | Full install package green. |
| `go test -count=1 -timeout 10m ./cmd/curator` | 0 | Fresh full command package green. |
| `go test -count=1 -timeout 10m ./internal/...` (first) | 1 | Setup failure: tracked `skill-go-testing-tools` submodule was uninitialized, so `internal/ui` replacement directory was absent; all other internal packages passed. |
| `git submodule update --init --recursive agents/skills/skill-go-testing-tools` | 0 | Materialized the pinned dependency. |
| `go test -count=1 -timeout 10m ./internal/...` (rerun) | 0 | Fresh full internal package set green. Together with `cmd/curator`, this is exactly `go list ./...`. |
| `go test -count=1 ./agents/...` | 1 | Expected command-shape error: submodule is a separate module and the root pattern matches no packages; not presented as a test pass. |
| `go test -timeout 10m ./...` | 143 | Extra unsplit run was terminated after exceeding the headless budget; not presented as green. The same root package set had already passed in bounded disjoint runs above. |
| `go test -p 2 -timeout 10m ./...` | 143 | Lower-contention extra retry also exceeded the shell budget and was terminated; it is not presented as green. The required root package set is supported by the two exact bounded commands above, as the headless instructions require. |
| final focused prompt/precheck tests | 0 | Includes the production-entry abort test added during self-review. |
| `go build ./...` | 0 | Root module compiles. |
| `go vet ./...` | 0 | Clean. |
| `golangci-lint run ./...` (final) | 0 | Clean. |
| `gofmt -l cmd internal` check | 0 | No files reported. |
| `git diff --check` | 0 | Clean. |

Logs are under `.temp/TASK-260825-3kb532/`. No files were staged or committed.

## Workspace anomaly

The accepted HTTPS dependency tasks were reviewer-accepted on the board, but
their repository bytes existed only as uncommitted source/test files in the
primary checkout while the managed Story worktrees remained clean at the base
commit. This task mechanically imported only the accepted files named by those
tasks' outcome evidence; unrelated primary-checkout board, CI, research, and
log changes were not copied. The anomaly and exact validation interpretation
are recorded in `LOGBOOK.md` entry 0052.
