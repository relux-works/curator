# TASK-260811-2gazym developer rework evidence

Status: prepared for independent review

Run: `RUN-260817-5cd5d5`

Authoritative run goal at the final validation checkpoint:
`GOAL-260817-d2482c` revision 1, resolved scope `TASK-260811-2gazym`.

Review policy: required

Date: 2026-08-18

## Reviewed finding input

This rework addresses R1 and R3 from
`TASK-260811-2gazym_review-verdict_RUN-260811-5b088d.md` and preserves the
reviewer's accepted R2 bounded-metadata implementation. No unrelated redesign,
commit, or staging operation was performed.

## R1: real selected-tool execution and production negatives

- The external-package positive still selects
  `curator-runtime-go-toolchain-v1`, admits the selected executable, and calls
  `AuthorizeSelectedAdapterExecution`. It now launches the exact absolute path
  returned by that authorization with `exec.CommandContext`; it performs no
  PATH lookup.
- Execution has a 10-second context and a 4096-byte combined output bound. The
  process-start counter increments only after `cmd.Start` succeeds. The exact
  output must equal `go version <role-evidence version> <role-evidence
  GOOS/GOARCH>`, binding observed execution back to the selected manifest.
- A same-entry-count dependency with different admitted bytes and manifest
  digest is rejected before launch, with the process-start count unchanged.
- Production recheck tests start from a selected/admitted controlled root, then
  add a real escaping symlink or a real FIFO. The real tree fingerprinter rejects
  both before any process starts.
- The local-output negative now creates an actual hard link from a pre-existing
  object and verifies `os.SameFile`. With no manager receipt issuer available,
  admission rejects, no admission token exists, and cache publication count
  remains zero. Production A08 authority stays unavailable until
  `TASK-260811-27xisf`.

## R3: complete BSD self-rehashed forgery evidence

- The BSD regular-member suite now self-rehashes mutations to extended-name
  size, extended-name hash, declared/member sizes, original name, canonical
  logical path, and coherently adjusted traversal accounting. Every mutation is
  rejected by `DecodeManifest`.
- The BSD `__.SYMDEF` metadata suite now performs the corresponding self-rehashed
  mutations to extended-name size/hash, original name, canonical metadata path,
  and coherently adjusted traversal accounting. Every mutation is rejected by
  `DecodeManifest`.
- Valid BSD regular and symbol-table manifests continue to round-trip.

## Preserved acceptance invariants

- The accepted conformance test still requires and passes exactly 182 unique
  exact-digest A/C/F/T/V cases, including F14 logical order independence with
  exact raw-payload manifest binding.
- `internal/artifactpolicy/conformance/corpus.go` remains byte-identical at
  SHA-256 `87a5cb6afb1c120cf75979cccd57fe2702c9a7dd74bee22dfa80418e1f26750e`.
- Compiled dependency handling remains globally fail-closed.
- `verified-binary-v1` and production A08 output authority remain unavailable.
- Kotlin/Gradle/Maven remain absent from `internal/artifactpolicy`.

## Files changed in this rework

- `internal/artifactpolicy/manager_external_test.go`
- `internal/artifactpolicy/toolchain_manager_unix_test.go` (new)
- `internal/artifactpolicy/reviewer_run_290cd4_test.go`

The worktree already contained the untracked artifactpolicy implementation and
unrelated board, research, spec, planning, and closuregraph state. Those were
preserved.

## Validation evidence

Every gate ran directly as a standalone process. No validation was piped
through `tee`; fingerprint pipelines enabled `pipefail`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| focused R1/R3 test command covering real Go execution, hardlink rejection, real symlink/FIFO rechecks, dependency drift, and both BSD forgery matrices | 0 | passed in 5.523s |
| `go test -count=1 ./internal/artifactpolicy -run '^TestReusableArtifactManifestV1ConformanceCorpus$'` | 0 | exact 182-case corpus passed in 1.040s |
| `go test -count=1 ./internal/artifactpolicy/...` | 0 | package passed in 38.607s; conformance subpackage has no test files |
| `go test -race -count=1 ./internal/artifactpolicy/...` | 0 | race suite passed in 241.315s |
| `go vet ./internal/artifactpolicy/...` | 0 | no findings |
| `go build ./internal/artifactpolicy/...` | 0 | package compiled |
| `gofmt -l internal/artifactpolicy` | 0 | no files listed |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./internal/artifactpolicy/...` | 0 | `0 issues.` |
| `go test -count=1 ./internal/buildsource ./internal/buildmeta ./internal/buildcache ./internal/godriver ./internal/buildrepo ./internal/install/...` | 0 | Go baseline passed; godriver 46.274s, buildrepo 20.166s, install 79.505s, atomicity 90.212s |
| `go test -count=1 ./...` | 0 | repository passed; cmd/curator 385.424s, artifactpolicy 153.964s, closuregraph 15.067s, godriver 69.811s, install 119.173s, atomicity 119.571s |
| `go vet ./...` | 0 | no findings |
| `go build ./...` | 0 | repository compiled |
| `gofmt -l cmd internal` | 0 | no files listed |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./...` | 0 | `0 issues.` |
| `git diff --check` | 0 | no tracked whitespace errors |
| Kotlin/Gradle/Maven exclusion check | 0 | no matches in `internal/artifactpolicy` |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

## Source stability

- Current artifactpolicy sorted path/content fingerprint:
  `e9243e1d753e71427a920b15244f08d9476f54bcf154df8751a91839accc49f9`.
- The repository lane covered 365 sorted `.go`, `go.mod`, and `go.sum` paths.
  Its fingerprint was
  `662e82e1396a989cc263baa223285956ffe643f80ca838f3f9db370c7a175461`
  immediately before the full test, immediately after it, and after full
  vet/build/format/lint.

No blocker or forced-fit constraint remains. The implementation is ready for a
new independent review.
