# TASK-260720-jrrgw9 — independent rc.5 full verifier results

Date: 2026-07-29
Role: tester
Verdict: return to development

## First actionable failure

The exact required primary macOS gate failed:

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
GOTMPDIR=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/verifier/gotmp \
go test -count=1 ./...
```

- Exit: **1**
- Elapsed: **601 seconds** from log birth/modify timestamps (`2026-07-29T14:09:37+0400` to `2026-07-29T14:19:38+0400`)
- Failing package: `github.com/relux-works/curator/cmd/curator`
- Package result: `FAIL .../cmd/curator 600.435s`
- Failure: `panic: test timed out after 10m0s`
- Test running when the package alarm fired: `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall (28s)`

The timeout was not changed, as required. This is an aggregate package-duration failure, not an assertion failure attributed to the named test: that test had run for only 28 seconds when the package-wide default 10-minute alarm fired. The development owner needs to make the exact uncached `go test -count=1 ./...` gate fit the unchanged default timeout, for example by reducing or safely restructuring the cumulative `cmd/curator` test workload. The verifier must not hide this by increasing the timeout.

Per the stop-on-first-actionable-failure instruction, `go test -race -count=1 ./...` and native Windows launcher execution were **not run** and are **not claimed**. The `ssh win` host was not mutated. Linux remained supplemental and non-gating; no operator-approved Go 1.25.5 GOROOT was available and no software was installed or downloaded.

## Provenance and exact delta

- Candidate: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
- Candidate HEAD: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Accepted integrated comparison worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree`
- Authoritative root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- `vectors/build-drivers.json` SHA-256: `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`
- Matrix SHA-256, both local and attached board copies: `3e4e2ee020841a9f45ce11c788f7617b8dd7ec2a64dfcace9fc968c8dbe7e9f2`

The matrix's nine latest rework-file hashes all matched exactly. A checksum comparison against the accepted integrated worktree showed a broader task-owned delta of **22 test files only**: 20 additions and two modified conformance tests (`internal/buildcache/conformance_test.go` and `internal/closure/conformance_test.go`). No product, schema, golden, registry, release-pin, or configuration delta appeared after excluding worktree metadata and ignored task/board directories. This is consistent with the matrix's test-only claim, although its provenance section enumerates only the latest nine-file rework subset rather than the complete 22-file task delta.

No candidate or authoritative-suite file was edited, staged, committed, published, or repinned by this verifier.

## Matrix and authoritative-suite audit

The focused executable barrier read the immutable authoritative root at runtime and passed all selected bindings:

- Build-driver positives: 8 published cases
- Rejections: 77 published cases across cache (16), compiler-directive (3), context (2), dependency-graph (14), execution-policy (2), filesystem (14), manifest (8), module (2), process (11), and toolchain (5)
- Toolchain identity: 12 cases
- Build-source identity: 10 cases
- Manager lifecycle: 2 launcher, 3 bootstrap, 3 upgrade, and 2 dry-run cases
- GC retention: all 5 roots in `gc-retains-roots`

The matrix's published launcher, bootstrap, upgrade, dry-run, cache, and GC names and counts match the immutable JSON. The focused run executed the dynamic consumers for identity, fixed environment and argv, policy, package graphs, manifest/filesystem rejections, stable driver/cache errors, build-source, cache outcomes, launcher process behavior, lifecycle, dry-run, GC, context, schema 6, and interop goldens.

The schema-7 external-repository cases remain routed rather than forced, as the matrix records. This verifier found no evidence that the candidate publishes the absent `build_repository_*` implementation surface.

## Gates and real exit codes

| Gate | Exit | Evidence |
| --- | ---: | --- |
| Tool readiness (`task-board`, Go, git, ssh, shasum, gofmt, df) | 0 | Go `go1.25.5 darwin/arm64`; task-board `0.23.0` |
| Initial no-process barrier | 0 | no matching Go/test process |
| Initial disk barrier | 0 | 23,385,112 KB available |
| Candidate/matrix hash and provenance checks | 0 | hashes above |
| Exact delta (`rsync -nrc --delete` with metadata/ignored exclusions) | 0 | 20 added tests, 2 modified tests |
| Exact 22-file `gofmt -l` | 0 | empty output |
| Pre-runtime `git diff --check` | 0 | empty output |
| Focused authoritative consumer test barrier, uncached | 0 | 35 seconds by log timestamps; all 12 packages `ok` |
| Post-focused no-process barrier | 0 | no matching Go/test process |
| Required `go test -count=1 ./...` | **1** | default 10-minute timeout in `cmd/curator` |
| Post-full no-process barrier | 0 | no matching Go/test process |
| Scoped `go vet` over 12 affected packages | 0 | empty output |
| Post-runtime exact 22-file `gofmt -l` | 0 | empty output |
| Post-runtime `git diff --check` | 0 | empty output |
| Final exact-delta recheck | 0 | unchanged 22-test-file delta |
| Final authoritative and matrix digest recheck | 0 | unchanged |
| Final no-process barrier | 0 | no matching Go/test process |
| GOTMPDIR removal and absence check | 0 | exact task-owned tree was 0 KB, removed with `rmdir`, absence verified |

Not run because the first required full gate failed:

- `go test -race -count=1 ./...`
- affected-package coverage measurement
- native Windows launcher test binary execution
- supplemental Linux execution

These are missing gates, not passes.

## Logs and hashes

- `verifier/focused-new-tests.log` — SHA-256 `f82cd003b3136ea1646a9bd3a88cad50ab157ff30d59cc7c8cce93cb038d97ab`
- `verifier/go-test-all.log` — SHA-256 `25af3516ca076c95777b8c01ca5f89a85c42df7e13ec5b804c6a28f7ca2c2385`
- `verifier/go-vet-scoped.log` — SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`

## Disk and cleanup

| Point | Available KB |
| --- | ---: |
| Initial | 23,385,112 |
| After focused barrier | 23,329,228 |
| After failed full gate | 23,229,732 |
| Final | 23,220,596 |

Free space never fell below the 12 GiB race-start threshold or the 8 GiB diagnostic stop threshold. Candidate size remained 3,676 KB and authoritative conformance-root size remained 2,068 KB. The only task-owned GOTMPDIR was empty after descendant termination, then removed exactly and verified absent. No matching Go/test process remained.

## Development handback

First required action: make the exact command `CURATOR_CONFORMANCE_ROOT=<immutable-root> go test -count=1 ./...` pass uncached without altering Go/package timeouts. After that, an independent verifier must rerun the full gate, then the full race gate only after a terminal-process and ≥12 GiB disk barrier, and then perform the native Windows launcher gate before review handoff.
