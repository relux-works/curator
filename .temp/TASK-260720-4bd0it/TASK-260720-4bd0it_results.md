# TASK-260720-4bd0it — install marker v2 and build currentness

## Provenance

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-4bd0it/worktree`.
- Exact detached base: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` (`origin/main`).
- Imported predecessor: byte-identical complete accepted product diff from `TASK-260720-6i3cya`; tracked diff SHA-256 `b318b469e6b980cd3cd55f51729a19113dfa32696f3f5199a451a5258e88f3b4`, untracked product-manifest SHA-256 `2e6ac3e6d7e72e9b07c51d16fc86151fbdce02931a0362b2f1651bc28bcfbd3a`.
- No blocked `TASK-260720-1zntv0` host-policy code, board/config, planning/research, diagram, cache, binary, alternate-index, or unrelated files were imported.
- Task-only product differences from the accepted predecessor are `internal/marker/marker.go`, `internal/marker/marker_test.go`, new `internal/marker/marker_v2_test.go`, and the legacy/v2 transition assertion in `internal/interop/golden_test.go`.
- Task-only file-manifest SHA-256: `c2d3094dc6fe9c4fcaadfe24bd7db7c42516d1013fa9e18e06d745f3e774fbf5`.
- Candidate rc.4 conformance suite was consumed only as non-release test input at suite SHA-256 `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`.

## Implementation

- Marker readers accept strict historical schema 1 (skill schemas 1–5) and strict schema 2 (skill schemas through 6), with schema-specific exact fields.
- Every `Write` mutation emits schema 2, required non-nil sorted sets, empty `build_roots`/`builds`, conditional `build_source`, exact `go-v1` per-command build records, and a terminal LF; invalid state is rejected before filesystem mutation.
- Schema-1 marshaling retains its historical wire shape, while rewriting a read schema-1 marker transitions it to canonical schema 2.
- `Current` preserves legacy currentness and adds optional `BuildCurrentness` callbacks for validated raw snapshot acquisition and protected cache inspection without importing installer orchestration.
- Compiled currentness fails closed for changed roots or static context/runtime exposure, missing/invalid/mutated raw snapshots, build-source drift (including package root marker bytes), missing input proof, unsupported or mismatched driver/input/key/target/toolchain/path, non-hit/untrusted/corrupt cache results, receipt-byte/hash drift, and artifact metadata drift.
- Installed-tree `hashing.ContentSHA256` behavior remains unchanged and continues to exclude the manager marker.

## Tests and validation

Passed:

- Candidate `go test ./internal/marker -count=1` covering all 14 authoritative install-marker-v2 schema cases, authoritative compiled marker writer round-trip, v1 schemas 1–5, source/context/cache/artifact drift, callback ordering, and marker-excluding hashing.
- Candidate `go test -race ./internal/marker -count=1 -cover` — 81.4% statement coverage.
- Candidate `go test ./internal/interop -run '^TestGoldenMarkerObject$' -count=1` — historical marker v1 remains readable and mutation output is v2.
- `make check` — repository `go vet`, full tests, and formatting gate.
- `go test -race ./... -count=1` — full race suite.
- `go build ./...` — native Darwin/arm64 build.
- `GOOS=linux GOARCH=amd64 go test -exec=/usr/bin/true ./...` — full Linux compile graph.
- `GOOS=windows GOARCH=amd64 go test -exec=/usr/bin/true ./...` — full Windows compile graph.
- `git diff --check` and `gofmt -l internal/marker internal/interop/golden_test.go` — clean.

`golangci-lint` is not installed on this host, so it was not run; the project-defined `make check` vet/test/gofmt gate passed. Native Windows execution was unavailable on this Darwin host; Windows production and test sources compile successfully and native runtime remains a CI gate.

## Logbook findings

- Optional candidate-wide interop still reaches the pre-existing third rc.4 `compiled-cache-miss-is-read-only` lifecycle vector while `internal/interop.TestManagerLifecycleVectors` models two dry-run cases. This is the already-recorded downstream `TASK-260720-jrrgw9` integration gap; marker-owned candidate cases and the legacy marker transition test pass, so it was not force-fit here.
- `task-board validate` still reports only the known 12 legacy `EPIC-260712-*` prose dependency references plus orphan `TASK-260713-7a9c1e/review.md`; no issue points to this task.
- No files were staged or committed.
