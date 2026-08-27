# TASK-260729-1zex8r tester collection-race regression

## Outcome

The digest-phase rework closes the previously reported detached-directory
handle regression, but the directory-scoped collection phase still diverges
from the legacy `fs.WalkDir` fail-closed behavior during a directory-to-file
replacement.

Added `TestToolchainWalkRejectsDirectoryReplacedByFile` to
`internal/godriver/fingerprint_equivalence_test.go`. The test is expected red
against the current product implementation and requires developer rework.

## Exact evidence

Previously reported digest-phase regression:

```text
go test -count=1 ./internal/godriver -run '^TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot$'
```

Real exit code: `0`.

```text
ok  	github.com/relux-works/curator/internal/godriver	0.551s
```

New collection-phase regression:

```text
go test -count=1 ./internal/godriver -run '^TestToolchainWalkRejectsDirectoryReplacedByFile$'
```

Real exit code: `1` (expected red because the product currently accepts the
replaced path).

```text
--- FAIL: TestToolchainWalkRejectsDirectoryReplacedByFile (0.00s)
    fingerprint_equivalence_test.go:854: scoped walk accepted a file in place of a listed directory: [{path:entry kind:70 link: info:0x140000a3040}]
FAIL
FAIL	github.com/relux-works/curator/internal/godriver	0.350s
FAIL
```

Formatting gates for the changed test file:

- `gofmt -d internal/godriver/fingerprint_equivalence_test.go` — exit `0`, no output.
- `git diff --check -- internal/godriver/fingerprint_equivalence_test.go` — exit `0`, no output.

## Why this is a legacy-equivalence failure

`fs.WalkDir` decides whether to descend from the `fs.DirEntry` returned by the
directory listing. Its callback may observe that the same root-relative path
has become a file, but after the callback it still attempts to read the path as
the directory that was listed and therefore fails closed.

`toolchainWalk.descend` instead overwrites the scoped directory metadata with a
root-relative `Lstat`, derives `record.kind = 'F'`, and uses that derived kind
to skip descent. The walk can then accept the replacement file as a leaf and
silently omit the listed directory's former descendants.

The deterministic test models the exact seam between the scoped directory
`Lstat` and the root-relative `Lstat`: the scoped handle reports the listed
directory while the root anchor reports the replacement file.

## Constraint and recommendation

This is a trust-boundary product defect, not a test-harness limitation. The
tester role does not authorize changing `internal/godriver/fingerprint.go`, and
adding a permissive assertion would legitimize the fail-open behavior.

Recommended rework: preserve the listed `DirEntry.IsDir()` decision separately
from the later canonical record classification. If an entry listed and scoped
as a directory is no longer a directory at the root-relative recheck, reject
the mutation before appending a leaf record. Retain the existing full-root file
open and `os.SameFile` check in the digest phase.

After rework, rerun the focused godriver suite, coverage, accepted go-v1
conformance/vector gates, build, vet, gofmt/diff gates, fingerprint benchmark,
and clean `cmd/curator` count-one timing. Those gates were intentionally not
continued after this expected-red trust failure.
