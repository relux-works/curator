# TASK-260729-1zex8r tester evidence — cycle 3

Date: 2026-07-29  
Role: tester  
Source worktree: `.temp/TASK-260729-1zex8r/worktree`  
Accepted integrated currentness source: `.temp/TASK-260720-1nlmvv/worktree`

## Outcome

The directory-scoped collection traversal is independently verified against the
preserved legacy traversal and is ready for review.

- The differential trust-boundary suite passes for canonical record ordering,
  record content, digest identity, diagnostic code and diagnostic detail.
- Both previously reported race regressions pass.
- The real host GOROOT produces the same 16,093 records and digest through both
  traversals.
- The affected fingerprint file has 87.1% statement coverage.
- Fingerprint latency improves from 1.559 s/op to 1.081 s/op (1.44x, 30.7%
  lower).
- The literal default-timeout `go test -count=1 ./cmd/curator` gate improves
  from 564.778s package time to 441.177s and passes below 480 seconds.

No product code was changed by this tester cycle. The owned differential tests
from the prior tester/developer cycles were preserved and rerun against the
reworked implementation.

## Scope and currentness audit

The rework source was compared against its task-imported baseline:

```text
rsync -rcn --delete --exclude='.git' --exclude='.task-board/' \
  --exclude='curator' --exclude='.temp/' --out-format='%i %n' \
  .temp/TASK-260729-1zex8r/before-worktree/ \
  .temp/TASK-260729-1zex8r/worktree/
```

Real exit code: **0**.

The only content deltas were:

- `internal/godriver/fingerprint.go`
- added `internal/godriver/fingerprint_equivalence_test.go`

All other reported godriver entries were timestamp-only. The refreshed exact
patch is `TASK-260729-1zex8r_fingerprint-cycle3.patch`, SHA-256:

```text
a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb
```

Two private copies were made from the accepted integrated currentness tree.
Before patching they were byte-identical by `rsync -rcn --delete` (exit **0**,
no content output). Applying the refreshed patch to the candidate exited **0**.
The post-apply comparison exited **0** and reported exactly the production
fingerprint file plus the added differential test. For like-for-like
fingerprint timing, an identical minimal benchmark-only test was then placed in
both disposable trees; their only production-code difference remained
`internal/godriver/fingerprint.go`.

## Trust-boundary regression gate

Command:

```text
go test -count=1 -run '^(TestFingerprintTraversalMatchesLegacyOnRepresentativeTrees|TestFingerprintTraversalMatchesLegacyOnFailClosedTrees|TestFingerprintTraversalMatchesLegacyOnErrorPrecedence|TestFingerprintTraversalMatchesLegacyOnRealToolchain|TestFingerprintIsStableAcrossRepeatedRuns|TestFingerprintDetectsMutationBetweenRuns|TestFingerprintCancellationStaysFailClosed|TestFingerprintDoesNotDescendLinkedDirectories|TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot|TestFingerprintDigestPhaseResolvesReplacedAncestors|TestToolchainWalkAnchorsDirectoryAndLinkMetadataAtTheRoot|TestToolchainWalkRejectsDirectoryReplacedByFile|TestToolchainWalkTakesDescentFromTheListedEntry|TestFingerprintRejectsFileRecordResolvingToADirectory|TestFingerprintReportsUnreadableDirectoryIdentically)$' -v ./internal/godriver
```

Real exit code: **0**. Package time: **3.641s**.

Coverage includes:

- representative flat, nested, deep, empty, Unicode and large-file trees;
- byte-for-byte digest and element-for-element canonical record equivalence;
- same-directory, parent-directory, linked-directory and link-to-link cases;
- absolute, escaping, dangling, deep-dangling and cyclic links;
- error precedence across siblings and depths;
- cancellation before collection, on an empty root, racing phases, and exactly
  at the collection/digest boundary;
- directory rename/replacement, ancestor replacement, file replacement,
  file-to-directory replacement and directory-to-file replacement;
- directory replacement by an in-root symlink;
- rename-away/rename-back ABA attempt;
- identical failure code and operator detail;
- no descent through a listed symlinked directory.

Real GOROOT evidence:

```text
GOROOT /Users/iv/.goenv/versions/1.25.5: 16093 records,
digest sha256:ea13c6bb11293e951ab9f189144a1f660cb2f398385109c0a3f7ad4875942191
```

## Validation gates

Every gate below ran as a standalone process without `tee`. Reported exit codes
are the real process exit codes.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| mutation/equivalence matrix | 0 | `logs/cycle3/mutation-equivalence.log` |
| `go test -count=1 ./internal/godriver` | 0 | 27.213s |
| conformance-root `go test -count=1 ./internal/godriver` | 0 | 30.866s |
| conformance/vector focused run | 0 | fingerprint framing and implementation vectors pass |
| `go build ./...` | 0 | empty log |
| `go vet ./...` | 0 | empty log |
| `gofmt -d` on both changed files | 0 | empty log |
| `git diff --check` | 0 | no output |
| `git apply --check --whitespace=error-all` refreshed patch | 0 | no output |
| affected coverage run | 0 | package 73.1% |
| `go tool cover -func=...` | 0 | fingerprint details below |

The accepted pinned conformance root is:

```text
.temp/TASK-260729-2kaopg/protocol-spec/conformance/v1
```

The vector run passed:

- `TestToolchainFramingMatchesRC4Vector`
- `TestFingerprintImplementationMatchesRC4ToolchainVector`

Ten tests skipped because this accepted pre-revision root publishes neither the
new execution-policy/build-driver vectors nor `expected/build-driver`; the raw
log records each reason. This matches the established baseline behavior and is
not reported as a pass for those absent artifacts.

Coverage:

```text
fingerprint.go statements: 135/155 = 87.1%
fingerprintToolchain       88.9%
collectToolchainRecords   100.0%
digestToolchainRecords     87.8%
toolchainWalk.descend      90.6%
readScopedDir              75.0%
```

## Performance evidence

Host Go/test-process barriers were captured immediately before, between and
after the decisive runs. Every barrier command exited **0**, and every barrier
log is zero bytes. No foreign Go compiler, linker, vet process or test binary
was observed at those checkpoints.

### Fingerprint A/B

Identical command in each private tree:

```text
go test -count=1 -run '^$' -bench '^BenchmarkFingerprintToolchainCycle3$' \
  -benchtime=3x ./internal/godriver
```

| Tree | Exit | ns/op | Package time | Wall |
| --- | ---: | ---: | ---: | ---: |
| baseline | 0 | 1,559,296,972 | 5.278s | 5.73s |
| candidate | 0 | 1,081,261,444 | 3.601s | 4.08s |

Result: **1.44x faster; 30.7% lower fingerprint latency**.

### cmd/curator default-timeout count-one A/B

Identical literal command in each private tree:

```text
go test -count=1 ./cmd/curator
```

No `-timeout` override was supplied.

| Tree | Exit | Package time | Wall |
| --- | ---: | ---: | ---: |
| baseline | 0 | 564.778s | 565.69s |
| candidate | 0 | 441.177s | 441.69s |

Result: candidate wall time is **124.00s lower (21.9%)** and **38.31s below**
the 480-second acceptance threshold. Candidate package time is 38.823s below
the threshold.

## Harness corrections reported honestly

Two preliminary measurement attempts failed before measuring product behavior:

1. The first baseline fingerprint benchmark exited **1** because copying the
   full differential test into the legacy tree referenced candidate-only helper
   functions. It was replaced with an identical minimal benchmark-only test in
   both disposable trees.
2. The first baseline cmd/curator timing exited **1** because the copy exclusion
   `curator` also omitted the `cmd/curator` directory. That directory was
   restored identically from the accepted currentness tree into both private
   copies before the decisive runs.

Neither failed setup attempt is counted as performance or passing evidence.
Their raw logs are retained in the evidence archive.

## Constraints honored

- no shared-cache clearing;
- no timeout increase;
- no assertion weakening;
- no product edit by this tester cycle;
- no accepted-worktree mutation;
- no host installation;
- no staging, commit, publication or pin change;
- no unrelated repository-wide or race suite;
- no claim that TASK-260729-2kaopg's two consecutive `go test -count=1 ./...`
  gates have run.

The fingerprint patch and test evidence are ready for independent review.
