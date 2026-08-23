# TASK-260822-27bvo4 - CI gate wiring for the symlink-launcher fix (RUN-260822-185ddd)

Scope of this run: the two review findings from RUN-260822-1f3d1d only. **No product
code was changed by this run** - `internal/godriver/identity.go` still carries the
same one-line fix (`physicalPath(absolute)` instead of
`filepath.EvalSymlinks(filepath.Clean(absolute))`) the reviewer already accepted, and
the test set is unchanged. What was missing was the CI wiring that makes the new
cases enforceable; that is what is verified below.

Tree: `.temp/TASK-260822-27bvo4/RUN-260822-507162`, branch
`task/TASK-260822-27bvo4-symlink-launcher-507162`, base `6a9b201`, uncommitted per
standing orders. Diff: 4 files modified + 2 new test files, 95 insertions.

## Finding 1 - unclassified skip reason (CI-fatal on a legitimate host condition)

`internal/godriver/identity_test.go:31` now skips with
`creating Windows symlink requires host support: %v` - the exact phrasing
`.github/ci/skip-classes.tsv` already classifies as `host-capability` / `allow`
(and the same phrasing `internal/buildcache/protection_windows_test.go` uses for
this situation). `skip-classes.tsv` is unchanged: no new row was needed.

All three skip reasons the new tests can print were matched against the class table
the way `platform-case-gate.sh` matches them:

| skip site | reason | class |
| --- | --- | --- |
| identity_test.go:31 | `creating Windows symlink requires host support: ...` | host-capability / allow |
| identity_test.go:175 | `this host cannot create a hard link: ...` | host-capability / allow (`this host cannot create`) |
| identity_windows_test.go:39 | `this host cannot create a junction os.Lstat reports as a non-directory` | host-capability / allow |
| identity_windows_test.go:115 | `this host cannot create a hard link: ...` | host-capability / allow |

None of them fired on either runner used below, so the classification is what
protects a host that lacks the privilege, not what the green runs depended on.

## Finding 2 - no platform-cases.tsv rows for the new identity cases

Nine rows added to `.github/ci/platform-cases.tsv` after the existing
`TestExecutableIdentityRejectsSubstitutionAndMutation` row (pure insertion; no
existing row was edited):

| case | must_run_on | skip_allowed_on | class |
| --- | --- | --- | --- |
| TestExecutableIdentityResolvesALauncherLink | darwin,windows | windows | host-capability |
| TestExecutableIdentityResolvesALauncherLink/* | - | windows | host-capability |
| TestExecutableIdentityRejectsSubstitutionThroughALauncherLink | darwin,windows | windows | host-capability |
| TestExecutableIdentityRejectsSubstitutionThroughALauncherLink/* | - | windows | host-capability |
| TestExecutableIdentityResolvesARealSymlinkedProcessLaunch | darwin,windows | windows | host-capability |
| TestBuildAcceptsAManagerStartedThroughALauncherLink | darwin,windows | windows | host-capability |
| TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink | darwin,windows | windows | host-capability |
| TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction | **windows** | windows | host-capability |
| TestExecutableIdentityStillRejectsSubstitutionBehindAJunction | **windows** | windows | host-capability |

The two junction rows are the ones the review asked for by name: the fix is a
Windows-only behaviour change, so a runner must be REQUIRED to execute the cases
that prove it. The `/*` subtest rows make a darwin subtest skip fatal (on darwin
both links are creatable, so a skip there is a defect, not a host limit) while
still tolerating the privilege-dependent Windows skip.

## Gates (real exit codes, each command run standalone, no pipes)

darwin/arm64, go1.25.5:

| gate | exit |
| --- | ---: |
| `gofmt -l ./cmd ./internal` (empty output) | 0 |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `GOOS=windows go vet ./internal/godriver/` | 0 |
| `GOOS=linux go vet ./internal/godriver/` | 0 |
| `golangci-lint run` (0 issues) | 0 |
| `make ledger-check` (**72 rows**, was 63) | 0 |
| `bash .github/ci/gate-selftest.sh` (75 passed, 0 failed) | 0 |
| `bash .github/ci/no-broad-suppression.sh` | 0 |
| `go test -json ./internal/godriver/ -count=1` | 0 |

Native Windows host (`ssh win`, DESKTOP-3PBO632, go1.25.5 windows/amd64), tree
shipped to `C:\curator-ci\dev185ddd` and removed afterwards:

| gate | exit |
| --- | ---: |
| `go build ./...` | 0 |
| `go test -json -count=1 ./internal/godriver/` (94.95s) | 0 |

## The rows were proved enforceable, not just consistent

`platform-case-gate.sh` was run against both real `go test -json` streams
(`CI_EXCLUDED_PKGS` = every other ledger package, `CI_DEFERRED_PKGS=internal/godriver`
because no conformance root is supplied here):

* **windows stream** (`CI_GATE_GOOS=windows`): all seven new cases observed
  `ok`, including both junction cases, with **zero skips among them**.
* **darwin stream**: the five darwin-required cases observed `ok`, zero skips
  among them.

Both gate runs exit 1 on exactly one line, identical on both platforms and
pre-existing:

```
FAIL  ledger case skipped for the wrong reason on <goos>: internal/godriver :: TestCandidateGoV1SourceAwareContract
      the ledger tolerates a root-content skip here; this one classified as root-unset
      reason: CURATOR_CONFORMANCE_ROOT is not set
```

That is the local no-root condition, not this change: the row and the case are
untouched by the diff (the ledger edit is a pure insertion), and CI supplies a
conformance root that makes the skip classify as `root-content`. Reported as red
rather than dressed up as green.

## Not run

* `go test ./...` - the host data volume sits at 99% (12 GiB free); earlier runs on
  this task died with `no space left on device` at full-suite parallelism. Package
  scope for this change is `internal/godriver` plus the CI tables; `go build ./...`
  and `go vet ./...` cover the rest of the module.
* Conformance gates - `CURATOR_CONFORMANCE_ROOT` unset (see the one gate failure
  above).

## Artifacts from this run

* `TASK-260822-27bvo4_ledger-check-RUN-185ddd.log`
* `TASK-260822-27bvo4_platform-case-gate-windows-RUN-185ddd.log`
* `TASK-260822-27bvo4_platform-case-gate-darwin-RUN-185ddd.log`
* `TASK-260822-27bvo4_skips-observed-windows-RUN-185ddd.tsv`
* `TASK-260822-27bvo4_skips-observed-darwin-RUN-185ddd.tsv`
