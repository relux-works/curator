# TASK-260811-2gazym R1-R3 developer rework evidence

Status: prepared for independent review

Run: `RUN-260811-a4d8a4`

Authoritative run goal at the final validation checkpoint:
`GOAL-260811-6671e8` revision 1, resolved scope
`TASK-260811-2gazym`

Review policy: required

Date: 2026-08-12

## Reviewed finding input

This rework addresses R1-R3 from
`TASK-260811-2gazym_review-verdict_RUN-260811-290cd4.md`. The reviewed
producer source fingerprint was
`d9a811e9f5ffadfcd4fda90b6a13b69ad3ba5a7a068e70fb1d12cd46a0bf7e1d`
and was reproduced before the first edit. No unexpected artifactpolicy path
was present.

## R1: production manager-owned toolchain authority

- Added the closed selector `curator-runtime-go-toolchain-v1`. Its public
  input is a selector ID and already admitted dependency tokens; there is no
  public root, executable path, fingerprint, seal, or trust boolean input.
- The central Go root is captured at package initialization. Selection binds
  the sorted exact dependency-manifest boundary, resolves the actual runtime
  Go executable, and fingerprints the complete root. The fingerprint includes
  portable paths, node kinds, portable mode bits, regular-file sizes and
  contents, and contained symlink targets; links that escape and special nodes
  reject.
- Admission reads the selected executable through `os.Root`, re-fingerprints
  the complete root, derives the sealed `ToolchainAuthorization` internally,
  and binds the resulting admission token to the private selection state.
  Immediate execution authorization rebinds all dependency admissions and
  re-fingerprints the root again before returning the executable path.
- Caller-supplied `Descriptor` fields remain audit labels. They cannot alter
  the central selector, root, executable, captured payload identity, role
  evidence, dependency-boundary digest, or execution authorization.
- The external-package positive test mutates caller-visible `go/build.Default`
  and `GOROOT`, yet admits and authorizes the real centrally selected Go
  executable. Arbitrary caller roots and copied pre-existing objects retain
  zero adapter starts/publications. Internal production-path regressions reject
  a changed dependency boundary and a post-admission executable-mode change.
- Positive local-output authorization remains unavailable until
  `TASK-260811-27xisf` provides protected causal receipts. No output issuer,
  allow flag, or dummy verifier was added.

## R2: full-path bounds before metadata materialization

- PAX local/global metadata is parsed record-by-record. Record and key framing
  is read with a bounded cursor; `path` and `linkpath` lengths are checked
  against the remaining full `container!/member` budget before their value is
  read or allocated. Other potentially large text values are streamed through
  UTF-8 validation.
- GNU long-name and long-link declared lengths are checked against that same
  full path budget before their value is materialized.
- GNU/COFF `ar` string tables are streamed and indexed without allocating the
  declared table. Every name is UTF-8/path validated in physical order,
  including names never referenced by another member. References must point
  to an exact indexed name start.
- BSD `#1/<len>` names preflight the full container path before the name read.
  Existing declared leaf/emitted limits still take precedence when crossed.
- Sparse 64 MiB cases keep each declared metadata member below the 256 MiB
  leaf limit while exceeding only the remaining full path budget. Local/global
  PAX, GNU long-name/link, GNU `ar` `//`, and BSD `#1` all reject with exact
  `max_path_bytes` evidence and allocate less than 16 MiB during inspection.
- A full service regression proves the new overlong-metadata diagnostic and
  unmanifested accounting seal and decode as canonical rejection evidence.

## R3: exact BSD extended-name accounting and evidence

- A successful BSD physical member charges `name bytes + logical member bytes`
  exactly once to leaf, emitted, and stream accounting. The logical member
  node still hashes and sizes only its member contents.
- The `archive-ar-v1` entry binds `declared_size`, `member_size`,
  `extended_name_size`, `extended_name_sha256`, the original encoded name,
  and the logical/container path relation. Codec accounting derives the
  physical emitted size from that evidence.
- Semantic decode requires exact arithmetic, name length/hash, original name,
  and logical path. It also handles BSD extended `__.SYMDEF` metadata through
  its canonical `$ar-metadata/bsd-symbol-table-NNN` path.
- Successful nested-source and BSD symbol-table regressions round-trip.
  Self-rehashed changes to extended-name size/hash, declared/member sizes, or
  original name all reject.

## Preserved conformance and scope

- `conformance.Cases()` still publishes exactly 182 unique exact-digest cases
  (`A=14`, `C=91`, `F=61`, `T=15`, `V=1`).
- `internal/artifactpolicy/conformance/corpus.go` is unchanged at SHA-256
  `87a5cb6afb1c120cf75979cccd57fe2702c9a7dd74bee22dfa80418e1f26750e`.
- Full raw/origin identity and the accepted F14 logical projection are
  unchanged.
- Compiled dependency classes remain globally fail-closed, Kotlin remains
  excluded, and `verified-binary-v1` remains unavailable.

## Files changed for RUN-260811-a4d8a4

- `internal/artifactpolicy/types.go`
- `internal/artifactpolicy/toolchain_manager.go` (new)
- `internal/artifactpolicy/manager_external_test.go`
- `internal/artifactpolicy/containers.go`
- `internal/artifactpolicy/semantics.go`
- `internal/artifactpolicy/limits.go`
- `internal/artifactpolicy/codec.go`
- `internal/artifactpolicy/reviewer_run_290cd4_test.go` (new)

The worktree already contained the untracked artifactpolicy implementation and
unrelated closuregraph/board changes. They were preserved.

## Final-source validation evidence

Every gate below ran directly as a standalone process. No gate was piped
through `tee`; pipeline-based fingerprint commands used `pipefail`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| focused R1-R3 command covering external positive/negatives, dependency/mode drift, all sparse path precedences, GNU unreferenced names, BSD accounting/metadata, and forged manifests | 0 | passed in 2.977s |
| `go test -count=1 ./internal/artifactpolicy -run '^TestReusableArtifactManifestV1ConformanceCorpus$'` | 0 | all 182 exact cases passed in 0.787s |
| `go test -count=1 ./internal/artifactpolicy/...` | 0 | package passed in 33.398s; conformance subpackage has no test files |
| `go test -race -count=1 ./internal/artifactpolicy/...` | 0 | full race suite passed in 236.683s |
| `go vet ./internal/artifactpolicy/...` | 0 | no findings |
| `go build ./internal/artifactpolicy/...` | 0 | package compiled |
| scoped gofmt cleanliness check | 0 | no artifactpolicy files listed |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./internal/artifactpolicy/...` | 0 | pinned v2.12.2, `0 issues.` |
| `go test -count=1 ./internal/buildsource ./internal/buildmeta ./internal/buildcache ./internal/godriver ./internal/buildrepo ./internal/install/...` | 0 | Go baseline passed; godriver 45.433s, buildrepo 19.569s, install 79.479s, atomicity 88.708s |
| `go test -count=1 ./...` | 0 | every repository package passed; `cmd/curator` 344.506s, artifactpolicy 124.357s, closuregraph 14.944s, godriver 59.455s, install 103.256s, atomicity 104.891s |
| `go vet ./...` | 0 | no findings |
| `go build ./...` | 0 | repository compiled |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./...` | 0 | `0 issues.`; one non-failing processor warning references a deleted stale `/private/tmp` integration worktree |
| full `cmd internal` gofmt cleanliness check | 0 | no files listed |
| `git diff --check` | 0 | no tracked whitespace errors |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

## Truthful repair-loop failures

These commands failed during implementation and are not represented as green
gates:

- Two early focused compile attempts exited 1 after the tar parser was changed:
  first because a removed payload read had left `readErr` undefined, then
  because a selector field was used with `:=`. Both compile errors were fixed.
- The first BSD extended symbol-table regression exited 1 because semantic
  path rederivation used the immediate synthetic-directory parent instead of
  the last container-chain element. Nested and metadata paths now use the
  exact containing archive.
- The first canonical overlong-PAX service regression exited 1 because the
  synthetic metadata diagnostic lacked the container in its chain. The chain
  and original encoded name are now bound, and canonical decode passes.
- The first pinned focused lint exited 1 with four G115 findings and one
  QF1003 finding. Signed parsing/textual size encoding and a tagged switch
  removed all five; the exact lint command now exits 0.
- The first focused compile after that lint repair exited 1 because a new
  `requiredInt64Fact` duplicated an existing helper. The duplicate was removed;
  focused, package, race, vet, build, and lint gates were rerun on final source.

One earlier long package command lost its shell-session result while being
polled and is intentionally not counted as evidence; later standalone package
gates above supply the authoritative exit 0 results.

## Source stability

- Current artifactpolicy sorted path/content fingerprint:
  `ca53bba924ed0cf8ecdf81be1a680cfe38bef04cfe9531ccb5adb01c02cba2d3`.
- The exclusive repository lane covered 364 sorted `.go`, `go.mod`, and
  `go.sum` paths. Its fingerprint was
  `dcf6064e5a87cf1f09237789fe0456e09f9d66f7f683a6b7c5821ad560d78910`
  immediately before full test, immediately after full test, and after full
  vet/build/lint.

No blocker or forced-fit constraint remains. This outcome is ready for a new
independent review.
