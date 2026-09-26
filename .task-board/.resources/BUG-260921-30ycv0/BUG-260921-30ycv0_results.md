# BUG-260921-30ycv0 results — gitops writeBlobs bare spawn error wrapped

## 1. Production-entry before/after (`install.Project` over a v1 git source, injected non-executable git)

Before (bare `cmd.Start()` return; reproduced via the mutant run below):

```text
fork/exec /var/folders/.../T/TestProjectWriteBlobsSpawnFailureIsWrapped758924147/004/git: permission denied
```

After (this change):

```text
git cat-file --batch failed in /var/folders/.../T/TestAfterMsgProbe1397248545/001/skill-a: fork/exec /var/folders/.../T/TestAfterMsgProbe1397248545/004/git: permission denied
```

The operation context names the failing spawn; the kernel cause (`permission denied`)
is preserved verbatim as the suffix and in the error chain (`%w`).

## 2. Change

- `internal/gitops/gitops.go` (`writeBlobs`): the `cmd.Start()` error is now
  `fmt.Errorf("git cat-file --batch failed in %s: %w", repo, err)` — the same
  operation-context shape the function's own `Wait`/abort paths already use.
  One line; no new refusal class; no retry; success path untouched.
- `CHANGELOG.md`: `## Unreleased` → `### Fixed` entry declaring the reworded failure.
- New tests:
  - `internal/gitops/writeblobs_spawn_test.go` — `TestWriteBlobsSpawnFailureIsWrapped`:
    in-package `writeBlobs` call with a PATH git shaped exactly as the brief
    prescribes (0755 script whose shebang interpreter is 0644). Asserts the
    `git cat-file --batch failed` context, the preserved `permission denied`
    cause, and `errors.Is(err, syscall.EACCES)`.
  - `internal/install/writeblobs_spawn_test.go` — `TestProjectWriteBlobsSpawnFailureIsWrapped`:
    production-entry row. `install.Project` over a v1 git source with a
    delegating shim git on PATH that, after serving the `ls-tree` listing,
    replaces its own shebang interpreter with a 0644 file — so the shim keeps
    resolving on PATH while the kernel refuses the very next exec, the
    writeBlobs `cat-file --batch` spawn, with EACCES. Asserts failed status,
    wrapped context, preserved cause, and that the trip fired (interp 0644).
- Caller error-class mapping: unchanged. Verified by inspection — no caller of
  `gitops.Extract` (closure, snapshot, contextstore, install) matches on the
  error text or with `errors.Is/As`; the old text survives as a substring and
  the chain is preserved via `%w`. No `git cat-file` string match exists
  outside `internal/gitops` except one `deadlock_test.go` assertion on the
  unrelated framing path (green, §4).
- Sanitization: the message adds the repository path, which the sibling
  messages in the same function (`ls-tree`, abort, `Wait`) already report —
  no new disclosure class.

## 3. Mutant table

| Mutant | `TestWriteBlobsSpawnFailureIsWrapped` (gitops) | `TestProjectWriteBlobsSpawnFailureIsWrapped` (install) |
|---|---|---|
| `writeBlobs`: restore bare `return err` from `cmd.Start()` | FAIL (`err = "fork/exec .../git: permission denied", want the wrapped operation context`) | FAIL (`errors = "fork/exec .../git: permission denied", want the wrapped cat-file operation context`) |

Both rows fail on the mutant and pass on the candidate (narrowing, not deletion:
the mutant exhibits the exact gate-observed bare message).

## 4. Verification (this candidate tree, exit codes real, `set -o pipefail`)

| Check | Result |
|---|---|
| `go test ./internal/gitops/ -count=1` (incl. byteexact goldens, deadlock tests, new row) | ok, exit 0 (9.4s) |
| `go test ./internal/install/ -run 'TestProjectWriteBlobsSpawnFailureIsWrapped\|TestGitFixture\|TestEndToEndInstall\|TestDryRunEffectBindingsSeeWhatARealOperationWrites\|TestDraftInstallGitPinnedSubtree'` | ok, exit 0 (5.9s); all PASS except the dry-run test, which SKIPs without `CURATOR_CONFORMANCE_ROOT` (ledger class `root-unset`, fires on CI only) |
| `CURATOR_CONFORMANCE_ROOT=<curator-spec>/conformance/v1 go test ./internal/install/ -run 'TestDryRunEffectBindingsSeeWhatARealOperationWrites\|TestAuthoritativeDryRunCasesMutateNothingPersistent'` (the gate-flake test + its suite) | ok, exit 0 (29.6s) |
| `go test ./internal/snapshot/ ./internal/contextstore/ -count=1` | ok, exit 0 |
| `go test ./internal/closure/ -count=1` | ok, exit 0 (18.3s) |
| `go vet ./internal/gitops/ ./internal/install/` | exit 0 |
| `golangci-lint run ./internal/gitops/... ./internal/install/...` | 0 issues, exit 0 |
| `gofmt -l` on touched dirs | clean |
| `go build ./...` | exit 0 |

Ratio line: no corpus row changed behaviour — zero goldens touched, zero
existing assertions edited; the two new rows are the only suite delta.
Windows lane: both new rows declare the sibling skip
(`executable-bit spawn refusal is exercised on the unix runners`, ledger
`platform-control /(is|are) exercised (on|by) /`); unverified here (darwin
host), will exercise on the unix gate lanes.

## 5. Rulings and bounds

- R1: done — spawn error wrapped with operation + repository context; `Wait`
  was already wrapped in the same shape (left as is); caller mapping unchanged.
- R2: done — no retry in product code; CHANGELOG Fixed entry present; legacy
  goldens green (§4).
- R3: done — production-entry row with injected non-executable git (§2);
  mutant fails both rows (§3); success-path rows existing and green (§4).
- R4 (spawn-observability seam): `gitops` has no logging/diagnostic seam —
  `run` and `writeBlobs` report only through returned errors, and there is no
  closure/install seam that could carry the fixture's binary/dir-stat and
  rlimit facts without printing environment or paths. Per the ruling, no
  product diagnostics were added; carrying fixture-grade forensics into
  product errors is recorded as a bound, not implemented.
- Bound: `writeBlobs`' `StdinPipe`/`StdoutPipe` errors stay bare. They are
  pre-spawn pipe setup (fd exhaustion), not the fork/exec spawn failure this
  bug covers, and no test can drive them deterministically; wrapping them
  without a killing row would be unproven churn.
- Notable finding while building the row: a shim that chmods *itself* 0644
  does NOT reproduce the flake through `exec.Command("git", ...)` — Go's PATH
  lookup skips non-executable entries and silently falls through to the real
  git. The refusal must hide behind a resolvable 0755 script (0644 shebang
  interpreter), which is why the brief prescribes exactly that shape; the
  install row trips on `ls-tree` and breaks its interpreter for the same reason.
