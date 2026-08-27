# TASK-260823-lk8hxy — Windows invalid-Unicode path validation

Fix-forward delivery. Supersedes PR 28 (`internal/fsunicode`), landed by the
previous run of this task, with PR 30.

## 1. Routing question: did PR 28 break the Windows Test lane?

No.

Failing test in run `32641695064` job `97199650872`:

```
internal/managerlock TestSubprocessBuildKeyDeduplicationAcrossProjects
  managerlock_test.go:509: independent build key helper = "blocked", want acquired
```

Three independent proofs it is unrelated:

| Proof | Result |
| --- | --- |
| Same commit `062d89b`, two Windows runs | `pull_request` run `32641695064` job `97199650872` FAILED; `workflow_dispatch` run `32641704975` job `97199678133` PASSED that exact test |
| `GOOS=windows go list -deps ./internal/managerlock` | no edge to `internal/fsunicode`, `internal/buildsource`, `internal/godriver`, `internal/identifiers` |
| Native Windows host (go1.25.5 windows/amd64) | `ok github.com/relux-works/curator/internal/managerlock 2.601s` |

It is a flake in a lock-contention subprocess test.

## 2. Root cause of the original vector failure

Both vectors publish the raw byte `0xFF` as the path payload
(`path_bytes_base64: "/w=="`).

Measured on a native Windows host — create a file with the given bytes, then
read the directory back:

| written bytes | read back | valid UTF-8 | round-trips |
| --- | --- | ---: | ---: |
| `FF` | `EF BF BD` (U+FFFD) | yes | no |
| `ED A0 80` (unpaired high surrogate, WTF-8) | `ED A0 80` | no | yes |
| `ED B0 80` (unpaired low surrogate, WTF-8) | `ED B0 80` | no | yes |
| `EF BF BD` (literal U+FFFD) | `EF BF BD` | yes | yes |

A Windows name is UTF-16. Go's `encodeWTF16` replaces the vector's ill-formed
UTF-8 with U+FFFD on the **write** path, so the probe created a member named
U+FFFD — perfectly valid Unicode — and the guard correctly accepted it. Go's
**read** path is lossless WTF-8 (go.dev/issue/59971), so an invalid scalar that
is actually on disk stays invalid UTF-8 in the string the guard sees.

The laundering was in the test harness, not in path validation. The guard was
already fail-closed.

## 3. Why PR 28 was wrong at the root

PR 28 added `internal/fsunicode`, whose Windows rule refused every string
containing U+FFFD, on the stated premise that "Go replaces an unpaired
surrogate with U+FFFD while converting that name to string". Row 2 of the table
above disproves that premise: unpaired surrogates round-trip intact.

Consequences of the rule as merged:

- it rejected literal U+FFFD names that no laundering could have produced,
  diverging Windows identity admission from POSIX;
- it put a test-harness workaround into the build-source and toolchain identity
  path, justified by a false platform claim.

## 4. What PR 30 does

- The probe materializes the vector in the spelling each host can carry: the
  vector's own bytes first, and on Windows an unpaired surrogate. It then
  asserts the directory really presents a name that is not valid Unicode before
  asserting anything about the guard, and skips as a host capability limit when
  no spelling survives (macOS/APFS rejects invalid UTF-8 names outright).
- `internal/fsunicode` is deleted and its two call sites restored. The product
  side is byte-identical to its state before PR 28 (verified with
  `git diff f761e50^1:<file> <file>` → empty for both files).

## 5. Validation

### Native Windows host, go1.25.5 windows/amd64, candidate root `curator-spec@859727b` (manifest `782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f`)

| Target | Result |
| --- | --- |
| `TestBuildSourceIdentityVectors/invalid-unicode-build-source-path` | PASS |
| `TestToolchainIdentityVectors/invalid-unicode-toolchain-path` | PASS |
| `internal/buildsource` (full package) | ok |
| `internal/identifiers` | ok |
| `internal/managerlock` | ok |
| `internal/godriver` | FAIL — see below |

`internal/godriver` still fails two cases on Windows against that root:

```
TestToolchainIdentityVectors/unsorted-directories-files-and-internal-link
  materialized link target = "..\\bin\\go", <nil>; want exact protocol target "../bin/go"
TestFixedEnvironmentAndFiveDirectArgvFormsVector/fixed_environment
  closed environment GOARCH = "amd64", the suite publishes "arm64"
```

Both reproduce identically on unmodified `origin/main` (`351db49`) on the same
host — a control run was executed specifically to establish this. They belong
to TASK-260823-czs1cx, not to this task.

### Local (darwin/arm64), same candidate root

| Command | Exit |
| --- | ---: |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `GOOS=windows go vet ./...` | 0 |
| `golangci-lint run` | 0 |
| `go test ./internal/buildsource ./internal/godriver ./internal/identifiers -count=1` | 0 |

On darwin both invalid-unicode cases SKIP: APFS refuses to create a name with
invalid UTF-8, which the skip-class ledger already classifies as
`host-capability` (`this host cannot create`). That matches the behaviour before
this change.
