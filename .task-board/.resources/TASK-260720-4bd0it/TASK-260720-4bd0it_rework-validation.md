# TASK-260720-4bd0it rework validation

## Scope

This rerun independently verified the reviewer-requested close-error rework in `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-4bd0it/worktree`.

- Base remains exact `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`.
- `internal/marker.currentBuilds` now consumes the validated snapshot token and explicitly passes `token.Close()` into `buildCurrentnessResult`.
- Close failures return `current=false` plus an operational error. Snapshot mutation without a close failure remains non-current without becoming an operational error.
- The regression test `TestBuildCurrentnessResultPropagatesSnapshotCloseFailure` covers the joined mutation and close-error result.
- No files were staged or committed.

## Fresh validation evidence

Passed on Darwin/arm64 with Go 1.25.5:

- `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test -race ./internal/marker -count=1 -cover` — pass, 81.5% statement coverage.
- `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test ./internal/interop -run '^TestGoldenMarkerObject$' -count=1` — pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./internal/marker/... ./internal/interop/...` — 0 issues.
- `make check` — pass (`go vet`, full tests, formatting gate).
- `go test -race ./... -count=1` — pass.
- `go build ./...` — pass.
- `GOOS=linux GOARCH=amd64 go test -exec=/usr/bin/true ./...` — Linux compile graph pass.
- `GOOS=windows GOARCH=amd64 go test -exec=/usr/bin/true ./...` — Windows compile graph pass.
- `git diff --check` and focused `gofmt -l` assertion — clean.

The standalone `golangci-lint` binary is unavailable, so lint ran through the reproducible `go run ...@latest` invocation used by review. Native Windows runtime execution is unavailable on the Darwin host; Windows production and test sources compiled successfully.
