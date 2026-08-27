# TASK-260720-1zntv0 implementation results

## Provenance and scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
- Exact detached base: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Imported predecessor: the complete accepted product state from `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-6i3cya/worktree`.
- A SHA-256 file-tree comparison excluding `internal/godriver`, Git metadata, and the initialized test-tool submodule found no byte differences from the accepted predecessor. No files outside `internal/godriver` were changed for this task.
- The rc.4 candidate was consumed only from `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree/conformance/v1`; it is test input, not release or pin evidence.

## Implementation

- Added the source-aware `Build` API around the accepted trusted `Session` and `buildsource.Token`.
- Added the exact fixed `go list` and `go build` vectors, canonical source-directory CWD, closed session environment, operation-private staging, and a build-tagged `HostPolicy` boundary carrying read-only roots, private writable roots, the closed executable graph, network denial, and deadline/output/artifact/disk/memory/process limits.
- The complete JSON package stream is decoded and validated before build. It requires one non-`DepOnly` canonical `main`, rejects incomplete/error/test results, validates standard inputs under fingerprinted GOROOT, and validates main-module plus vendored package/module/source/embed inputs under the build root.
- Added rejection of workspace/toolchain directives, missing/inconsistent vendor state, cgo/native/SWIG inputs, every syso, non-standard assembly, escaped or linked input, active `//go:cgo_import_dynamic`, active generator directives, and `default.pgo`.
- Both children are bracketed by frozen-source and exact-toolchain rechecks. Failed staging is removed.
- Artifact verification requires exactly one bounded regular non-reparse/non-hardlinked manager-derived output, applies manager executable permissions before hashing, returns path/size/SHA-256 metadata, and never starts the output.
- Combined stdout/stderr now share one bound in `OSExecutor`.

## Tests and evidence

- `go test ./internal/godriver -count=1` — pass.
- `go test ./internal/godriver -coverprofile=...` — pass, 80.2% statement coverage.
- `go test -race ./internal/godriver -count=1` — pass after the last production edits.
- `make check` (`go vet ./...` plus `go test ./...`) — pass after the last production edits.
- `go test -race ./...` — pass across the imported product tree.
- `go build ./...` — pass.
- `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test ./internal/godriver -count=1` — pass against rc.4 vector identity and rejection vocabulary.
- `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run 'TestCandidateGoV1SourceAwareContract|TestRealGoV1VendoredBuildIsBoundedAndNotLaunched' -count=1 -v` — pass with the authoritative standard-library plus vendored transitive/embed fixture and the installed Go 1.25.5 toolchain. The host-policy observer recorded exactly one list and one build call; no artifact call occurred.
- `GOOS=windows GOARCH=amd64 go test -c ... ./internal/godriver` and `GOOS=linux GOARCH=amd64 go test -c ... ./internal/godriver` — pass. Native Windows/Linux runtime evidence is unavailable on this macOS host.
- `git diff --check` and `gofmt -l internal/godriver` — clean.
- `golangci-lint` is not installed; repository lint-equivalent evidence is `go vet ./...`, formatting, focused/full tests, race, and native/cross builds.

Tests explicitly prove exact argv/environment/CWD, complete root and transitive graph rejection clusters, source-readonly and offline network/module/VCS settings, host-policy limit/root/executable interactions, source mutation through the last child, combined output/deadline/disk/artifact limits, regular/symlink/hardlink output handling, cleanup on failure, and the absence of built-output launch.

## Known unrelated candidate-suite anomaly

`CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test ./...` passes every package except the pre-existing `internal/interop` `TestManagerLifecycleVectors` reader, which still expects two dry-run vectors while rc.4 adds `compiled-cache-miss-is-read-only` as a third. Normal full tests pass, task-owned rc.4 consumers pass, and downstream lifecycle integration owns that reconciliation. This is the same accepted predecessor anomaly already recorded on upstream tasks.
