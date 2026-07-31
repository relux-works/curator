# TASK-260720-3mrm4z implementation evidence

## Provenance and integration boundary

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3mrm4z/worktree`
- Exact base: `origin/main` at `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Accepted predecessor: the complete product diff from `TASK-260720-256kj1` was imported before task work. Byte-for-byte comparison confirms that `go.mod`, `internal/buildsource`, and the accepted `internal/snapshot` changes remain identical to `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-256kj1/worktree`.
- Candidate conformance root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree/conformance/v1`
- Candidate manifest SHA-256: `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`
- The candidate suite was consumed only as explicit vector input. It is not release, pin, or provenance evidence.
- No task-board files/config, planning/research files, diagrams, binaries, caches, or unrelated shared-checkout files were imported into the product worktree.

## Task-only implementation

- Added `internal/protocoljson/ccj.go` with reusable CCJ-1 canonical encoding, exact byte comparison, strict canonical validation, and strict decoding with retained integer precision.
- Migrated `internal/registry.CanonicalBytesChecked` to the shared encoder while preserving the registry-specific top-level `sig` exclusion.
- Added `internal/buildmeta/models.go` and `internal/buildmeta/codec.go` for the complete logical go-v1 input, build-source identity, native target/tuning, toolchain identity, fixed directive policy, platform-derived artifact, receipt, cache key, and receipt hash.
- Added `internal/buildmeta/buildmeta_test.go` and expanded `internal/protocoljson/json_test.go`.
- `internal/buildcache` was not created or modified. The implementation does not inspect ownership, publish cache files, invoke Go as product behavior, or write install markers.

## Exact identities

- Authoritative build-input bytes: 830 exact CCJ-1 bytes; raw SHA-256 `3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`.
- Logical cache key: `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`.
- Authoritative receipt bytes: 1081 exact CCJ-1 bytes; raw SHA-256 `750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`.
- Receipt consistency identifier: `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`.
- Receipt hashes are documented and tested as deterministic corruption/currentness metadata, not signatures, attestations, authorization tokens, or proof of provenance.

## Strict rejection coverage

- Duplicate keys, non-integer numbers, negative zero, unsafe integers, invalid UTF-8/surrogates, a BOM, alternate escapes, member reordering, leading/trailing/pretty whitespace, and terminal LF.
- Every required input, build-source, target, toolchain, policy, receipt, and artifact field when absent.
- Unknown fields at all build metadata object levels, including timestamp/mode-style additions.
- Unsupported receipt, driver, build-source, and toolchain versions/algorithms.
- Absolute or nonportable paths, source directories outside the build root, malformed hashes, wrong cache keys, wrong derived artifact paths, and artifact sizes outside the non-negative safe integer range.
- Full expected input comparison covers build root, source directory, source identity, command, target and tuning, toolchain version/content identity, and the closed directive policy.

## Verification

- `make check` — PASS (`go vet ./...`, `go test ./...`, repository-wide gofmt check).
- `go build ./...` — PASS.
- `GOOS=windows GOARCH=amd64 go test ./... -run '^$' -exec=true` — PASS.
- `GOOS=linux GOARCH=amd64 go test ./... -run '^$' -exec=true` — PASS.
- Candidate-focused `go test -count=1 ./internal/protocoljson ./internal/buildmeta ./internal/buildsource ./internal/registry` — PASS.
- Candidate registry interop `TestCanonicalJSONVectors|TestGoldenRegistryObjects` — PASS.
- Focused race run for `protocoljson`, `buildmeta`, `registry`, `buildsource`, and `snapshot` — PASS.
- Coverage: `internal/protocoljson` 90.2%; `internal/buildmeta` 88.8%.
- `git diff --check` — PASS; `gofmt -l .` — empty.
- Accepted predecessor product files compared byte-for-byte — PASS.

Logs are retained under `.temp/TASK-260720-3mrm4z/`: `make-check.log`, `go-build.log`, `windows-compile.log`, `linux-compile.log`, `focused-race.log`, `candidate-focused-final.log`, `registry-vectors-final.log`, and `coverage-final.log`.

## Anomaly recorded

The fresh exact-base worktree initially lacked its registered `agents/skills/skill-go-testing-tools` submodule checkout, so the first standalone `go vet` could not resolve the local replacement. Initializing the submodule at its recorded commit `21585d0e937cae47e54a788d8ae36b1780eae47f` restored the intended repository environment. The subsequent full `make check`, including `go vet`, passed. This changed no product diff.
