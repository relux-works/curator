# TASK-260720-6i3cya — trusted Go toolchain session results

## Provenance

- Implementation worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-6i3cya/worktree`.
- Exact detached base: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` (`origin/main`).
- Imported predecessor: complete accepted product diff from `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-29hi1h/worktree`; no board/config, planning/research, diagram, cache, binary, or unrelated files were imported.
- Imported tracked-diff SHA-256 outside `internal/godriver`: `b318b469e6b980cd3cd55f51729a19113dfa32696f3f5199a451a5258e88f3b4` in both worktrees.
- Imported untracked-product manifest SHA-256 outside `internal/godriver`: `371df088038f9c11911731e62f8098c73786229328b70bd3b10bee379ea01706` in both worktrees.
- Task-only `internal/godriver` file-manifest SHA-256: `198855ec511a7e2d6ccdc11b5f0fd1f174f91936b0c8fc13888aa723ebee1a10`.
- Candidate rc.4 conformance root was used only as non-release vector input; manifest SHA-256: `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`.

## Implementation

Created `internal/godriver` with:

- absolute selection precedence `CURATOR_GO`, `GOROOT/bin/go`, then `runtime.GOROOT()/bin/go`, with no `exec.LookPath` or user/project `PATH` lookup;
- physical root/link resolution, repository/runtime-root exclusions, regular executable/native-image checks that reject wrappers, and exact `<GOROOT>/bin/go[.exe]` binding;
- a narrow injectable direct-process executor with closed stdin, per-command deadline, bounded stdout/stderr, no shell, and stable diagnostic codes;
- operation-private telemetry/config/home/cache/temp roots and an empty working directory plus empty manager-owned `PATH` directory;
- exactly one each of `go telemetry off`, `go version`, and fixed `go env -json ...`, with strict duplicate/unknown/missing JSON-field rejection;
- native host/target/version equality, private physical `GOTELEMETRYDIR`, telemetry-off verification, Go 1.23 floor, and the Curator rc.4-tested Go 1.25 family allowlist;
- one applicable architecture tuning value and the fixed clean operation environment, including empty `GOFLAGS`, `GOPRIVATE`, and `GOEXPERIMENT` plus offline/local-toolchain/linker controls;
- exact `curator-go-toolchain-v1` directory/file/link/version framing, unsigned bytewise path ordering, LF/CRLF version normalization, safe relative in-tree links, special/duplicate/invalid-UTF-8 rejection, and mode/timestamp exclusion;
- session-close re-fingerprinting through the last child boundary, mutation failure, cleanup on success/failure/drift, and a non-memoizing dry-run `Probe` API.

No package source is inspected and this package does not run `go list` or `go build`.

## Tests and validation

Passed:

- `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test ./internal/godriver -count=1 -cover` — 78.5% statement coverage; authoritative digest `sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e` and candidate artifacts pass.
- `CURATOR_REAL_GO_TEST=1 go test ./internal/godriver -run TestRealTrustedGoProbe -count=1` — native Go 1.25.5 Darwin/arm64 probe and cleanup pass.
- `make check` — full committed-suite `go vet`, tests, and gofmt gate pass.
- `go test -race ./...` — full repository race suite passes.
- `go build ./...` — native build passes.
- `GOOS=windows GOARCH=amd64 go test -exec=/usr/bin/true ./...` — full Windows compile gate passes.
- `GOOS=linux GOARCH=amd64 go test -exec=/usr/bin/true ./...` — full Linux compile gate passes.
- `git diff --check` and `gofmt -l internal/godriver` — clean.

`golangci-lint` is not installed on this host, so it was not run; the project-defined `make check` vet/test/gofmt gate passed. Native Windows execution was not available on this macOS host; Windows production/test sources and the full repository test graph compile successfully and CI remains the native runtime gate.

## Logbook findings

- The optional candidate-wide `CURATOR_CONFORMANCE_ROOT=... make check` reaches the new rc.4 third manager dry-run vector, but the pre-existing `internal/interop.TestManagerLifecycleVectors` still asserts exactly two dry-run cases. The task-owned rc.4 `godriver` vector consumer passes. This mismatch belongs to dedicated end-to-end conformance task `TASK-260720-jrrgw9` and was not force-fit into this package-owned delta.
- `task-board validate` continues to report the known 12 legacy `EPIC-260712-*` prose dependency references and orphan `TASK-260713-7a9c1e/review.md`; no issue points to this task.
