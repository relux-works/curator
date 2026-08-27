# TASK-260729-1zex8r tester mutation regression

## Outcome

The directory-scoped implementation does not preserve the legacy fail-closed
behavior when a cached directory is renamed out of GOROOT and replaced at the
same canonical path. The new traversal can continue opening later files through
the detached directory handle and hash bytes that are no longer present at the
canonical path.

This is a trust-boundary regression against acceptance criteria 1 and 2. The
performance and release gates were not continued because a faster result cannot
make the candidate acceptable while this expected-red regression remains.

## Scoped-delta preflight

Compared
`.temp/TASK-260729-2kaopg/worktree` with
`.temp/TASK-260729-1zex8r/worktree` using:

```text
rsync -rcn --delete --exclude='.git/' --exclude='logs/' --exclude='*.log' \
  --out-format='%i %n%L' BASELINE/ CANDIDATE/
```

Real exit code: `0`.

Ignoring generated `curator.test`, task-local `.temp/`, `.git` metadata, and an
unrelated vendor-file mtime, the candidate differs from the preserved baseline
only by:

- `internal/godriver/fingerprint.go`
- `internal/godriver/fingerprint_equivalence_test.go`

The board-attached producer test and the pre-tester worktree test had identical
SHA-256:

```text
dd356703c42edc01e67eb3b179b8d0b9fda206316fff599954173b3d61258952
```

## Regression

Added
`TestScopedDirsDoesNotReadFromAReplacedDirectory` to
`internal/godriver/fingerprint_equivalence_test.go`.

The test:

1. opens and caches the scoped handle for `pkg`;
2. renames `pkg` to `detached`;
3. creates a replacement `pkg` with different bytes;
4. proves a legacy root-relative `Open("pkg/second")` reads replacement bytes;
5. requires the scoped lookup either to read the replacement or fail closed.

Validation command:

```text
go test -count=1 ./internal/godriver \
  -run '^TestScopedDirsDoesNotReadFromAReplacedDirectory$'
```

Real exit code: `1` (expected red).

Exact failure:

```text
--- FAIL: TestScopedDirsDoesNotReadFromAReplacedDirectory (0.00s)
    fingerprint_equivalence_test.go:667: scoped lookup read detached bytes "old-second" after directory replacement
FAIL
FAIL	github.com/relux-works/curator/internal/godriver	0.416s
FAIL
```

Formatting checks:

- `gofmt -d internal/godriver/fingerprint_equivalence_test.go` — exit `0`,
  no output.
- `git diff --check -- internal/godriver/fingerprint_equivalence_test.go` —
  exit `0`, no output.

## Constraint

`scopedDirs.open` treats a matching cached component string as sufficient proof
that the cached handle still represents that path. A directory handle remains
usable after its directory is renamed, so string equality does not prove
current path binding. The subsequent file-level `os.SameFile` check compares
against the original file in the detached directory and therefore also passes.

The tester continuation explicitly forbids product edits. Fixing this requires a
new product design for validating reused directory handles without losing the
required performance improvement.

## Options

1. Revalidate the cached ancestor chain against GOROOT at each use. This most
   directly preserves legacy path binding, but repeated component resolution
   may consume part of the latency gain.
2. Add a directory-identity verification pass before returning the fingerprint.
   This can retain most of the speedup, but must prove that it closes the same
   race windows and preserves diagnostic code/detail rather than merely moving
   them.
3. Stop reusing directory handles during the hashing phase. This preserves the
   legacy behavior but abandons the main performance optimization.

Recommended: route to a godriver developer to implement option 1 or a formally
equivalent validation scheme, with deterministic end-to-end mutation hooks that
compare the old and new traversal at the mutation boundary. Then rerun
equivalence, conformance/vector, performance, and default-timeout gates.

## Required rework decision

Authorize product rework despite the tester's no-product-edit directive and
choose the path-binding validation strategy. Until then, the task cannot satisfy
the trust/race acceptance criteria or truthfully hand off to review.
