# TASK-260825-1lausy — per-repository HTTPS resolution

## Delivered

- Added `internal/install/buildhttps.go` with per-effective-repository HTTPS
  resolution: covering run-wide override, longest `build_https` scope, then
  anonymous HTTPS.
- Added `CURATOR_BUILD_HTTPS_TOKEN` capture and the optional exact-host pin
  `CURATOR_BUILD_HTTPS_HOST`. A repository outside the pin resolves as though
  the override were absent, including scoped credentials or anonymous fallback.
- Captured all configured `token_env` values at process entry and materialized
  `git-credentials` and `keyring` selections through `internal/gitcred`.
- Added redacted diagnostic formatting for captured and resolved secret-bearing
  types. `%v`, `%+v`, and `%#v` are covered by tests and never render secrets.
- Added source-specific fail-closed diagnostics and remedies for missing
  `token_env`, Git host credentials, and manager-namespaced keyring material.
- Bound resolved HTTPS credentials and provenance into `externalPlan`; actual
  AskPass/broker delivery remains the explicitly separate
  `TASK-260825-3n4bjj`.
- Wired production external dependencies to capture HTTPS selection at process
  entry.

## Tests

`internal/install/buildhttps_test.go` covers:

- run-wide precedence and unpinned all-host coverage;
- exact host pin with scoped and anonymous behavior on non-covered hosts;
- longest segment-aware scope;
- anonymous fallback;
- effective transport skip for SSH and local substitutions;
- capture immutability and diagnostic redaction;
- successful `token_env`, Git-host, and keyring materialization;
- all three missing-material remedies;
- resolved material/provenance carried by the external install plan before
  acquisition.

Focused coverage reports 100% for the resolver, configured-source materializer,
transport predicate, key/host helpers, and provenance formatter; capture is 75%
(only nil-input defensive branches are uncovered).

## Validation evidence

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -coverprofile=... ./internal/install -run 'Test(BuildHTTPS|CaptureBuildHTTPS)'` | 0 | focused behavior and coverage green |
| `go test -timeout 30m -count=1 ./...` (stable retry) | 0 | all packages green |
| `go build ./...` (final) | 0 | build green |
| `go vet ./...` | 0 | vet green |
| `golangci-lint run ./...` (final) | 0 | lint green |
| scoped `git diff --check` + `gofmt -l` | 0 | no findings in task files |

The first full-suite attempt exited 1 during a concurrent shared-checkout write:
`cmd/curator/main.go` contained the neighboring CLI task's new import/dispatch
before its function body existed. The unchanged command was rerun after that
write completed and exited 0. Both logs are preserved as
`go-test-all.log` and `go-test-all-02.log`; `cmd-compile.log` records the
intermediate compile-only confirmation at exit 0.

## Files attributable to this task

- `internal/install/buildhttps.go` (new)
- `internal/install/buildhttps_test.go` (new)
- `internal/install/external.go`
- `cmd/curator/main.go` (one production capture assignment; the surrounding
  `build-https` CLI delta belongs to another concurrent task)
- `LOGBOOK.md` (entry 0055)
