# TASK-260822-27bvo4 — review of the gate-wiring rework (RUN-260822-abf447)

**Verdict: accepted.** Both findings from the RUN-260822-1f3d1d review are closed.
Every claim below was reproduced by me from a real run — no implementer log was
read back as evidence. The reviewed tree was not modified: it still shows the
same 4 modified files + 2 new test files, 95 insertions.

Reviewed tree: `.temp/TASK-260822-27bvo4/RUN-260822-507162`,
branch `task/TASK-260822-27bvo4-symlink-launcher-507162`, base `6a9b201`.

## Scope of this cycle

The previous review accepted the product fix and the test set and asked for two
mechanical gate-wiring fixes. This rework changed no product code — `identity.go`
still carries the same one-line `physicalPath(absolute)` fix and the doc comment,
and the test set is unchanged. I did not re-litigate either; I re-confirmed the
fix is intact and concentrated on whether the two gate rows actually bite.

## Finding 1 (unclassified skip reason) — CLOSED

`internal/godriver/identity_test.go:31` now skips with
`creating Windows symlink requires host support: %v`, which is the exact row
`skip-classes.tsv` already carries as `host-capability / allow` and the phrasing
`internal/buildcache/protection_windows_test.go` uses for the same situation.
No new `skip-classes.tsv` row was needed and none was added.

I matched every skip reason the new tests can reach against the table the way
`platform-case-gate.sh` matches them:

| site | reason | class |
| --- | --- | --- |
| identity_test.go:31 | creating Windows symlink requires host support | host-capability |
| identity_test.go:175 | this host cannot create a hard link | host-capability |
| identity_windows_test.go:39 | this host cannot create a junction … | host-capability |
| identity_windows_test.go:115 | this host cannot create a hard link | host-capability |
| selection_windows_test.go:83 (`makeTestJunction`, pre-existing) | this host cannot create a directory junction | host-capability |

None of them fired on either runner below, so the classification protects a host
that lacks `SeCreateSymbolicLinkPrivilege` rather than propping up a green run.

## Finding 2 (no ledger rows) — CLOSED

Nine rows added to `platform-cases.tsv` as a pure insertion after the existing
`TestExecutableIdentityRejectsSubstitutionAndMutation` row; no existing row was
edited. `make ledger-check`: **exit 0, 72 rows** (was 63), and all seven new case
names resolve against the per-GOOS build file sets.

The rework deviates from the rows I suggested last cycle: it sets
`skip_allowed_on=windows / host-capability` where I wrote `-`. That deviation is
correct, not a shortcut. Four of the five shared cases genuinely skip at the
*parent* level on a Windows host without the symlink privilege
(`RejectsSubstitutionThroughALauncherLink` skips after its subtest loop;
`ResolvesARealSymlinkedProcessLaunch`, `BuildAccepts…` and `WorkerProves…` skip at
the top), so `-` would have re-created exactly the CI-fatal condition Finding 1
was about. The `must=X / skip=X / host-capability` shape is already the repo's
convention for this — `internal/buildcache` uses it for two reparse-point cases
and `internal/scopes` for one. My ask was that *some runner must be required to
execute the Windows case by name*, and that is satisfied.

The two `Parent/*` rows are the piece that makes it strict: a subtest skip on
darwin is fatal while the same skip on Windows is tolerated.

### Rows proved enforceable, not merely consistent

I drove `platform-case-gate.sh` against real `go test -json` streams from both
runners (`CI_EXCLUDED_PKGS` = every other ledger package,
`CI_DEFERRED_PKGS=internal/godriver`, no conformance root here), plus two
adversarial mutations of the darwin stream:

| check | result |
| --- | --- |
| darwin stream, all 5 darwin-visible cases | `ok`, 0 skips among them |
| windows stream, all 7 cases incl. both junction cases | `ok`, 0 skips among them |
| darwin stream with `ResolvesALauncherLink` renamed | **FAIL `required case never ran on darwin`** — by name |
| darwin stream with one subtest turned into a classified skip | **FAIL `ledger case skipped where the ledger does not tolerate it (darwin)`** |
| same windows stream against the **base** `platform-cases.tsv` | identical single failure — the new rows add no new red |

A rename, a deletion or a `-run` filter matching nothing now fails by name, and a
darwin subtest skip is fatal. That is the regression trap Finding 2 asked for.

## Gates I ran (real exit codes, each standalone)

darwin/arm64, go1.25.5:

| gate | exit |
| --- | ---: |
| `gofmt -l ./cmd ./internal` | 0 (empty) |
| `go vet ./internal/godriver/` | 0 |
| `GOOS=windows go vet ./internal/godriver/` | 0 |
| `golangci-lint run` | 0 (0 issues) |
| `go test -json ./internal/godriver/ -count=1` | 0 — 296 pass, 0 fail, 18 skip |
| `make ledger-check` | 0 (72 rows) |
| `.github/ci/gate-selftest.sh` | 0 (75 passed, 0 failed) |
| `.github/ci/no-broad-suppression.sh` | 0 |
| `platform-case-gate.sh` (darwin stream) | 1 — see below |

Native Windows host `ssh win` (DESKTOP-3PBO632, go1.25.5 windows/amd64), tree
shipped to `C:\curator-ci\rev-abf447` and removed afterwards:

| gate | exit |
| --- | ---: |
| `go build ./...` | 0 |
| `go test -json -count=1 ./internal/godriver/` | 0 — 297 pass, 0 fail, 17 skip |
| `platform-case-gate.sh` (`CI_GATE_GOOS=windows`) | 1 — see below |

**The one gate red, reported as red:** both `platform-case-gate.sh` runs fail on
exactly one line, identical on both platforms —
`TestCandidateGoV1SourceAwareContract` skips `root-unset` where its (untouched)
row tolerates `root-content`, purely because `CURATOR_CONFORMANCE_ROOT` is unset
locally. I ran the same Windows stream against the **base** ledger from `HEAD`
and got the identical single failure, so it is pre-existing and independent of
this change.

**Not run:** `go test ./...`. The host data volume sits at 99% (~10 GiB free) and
earlier runs on this task died with `no space left on device` at full-suite
parallelism. `go build ./...` and `go vet ./...` cover the rest of the module, and
the change touches only `internal/godriver` plus a CI table, so no other package
can be affected. Conformance gates not run (`CURATOR_CONFORMANCE_ROOT` unset).

## Fix re-confirmed intact

`identity.go:56` canonicalizes with `physicalPath(absolute)`. On unix
`physicalPath` is `filepath.EvalSymlinks(filepath.Clean(path))` verbatim
(`platform_unix.go:16`), so the change is bit-identical there; on Windows it is
`GetFinalPathNameByHandle`, which follows a junction where `EvalSymlinks` does
not. The substitution battery in `readExecutableIdentity` is untouched and now
applies to the physical file, and `Verify()` deliberately does *not*
re-canonicalize — it re-`Lstat`s the recorded physical path, which is why
swapping that file for a link is still caught. Resolution first, rejection
second, exactly as `profiles/manager.md` orders it.

## Minor, non-blocking

`identity_test.go:62` declares `skipWin bool` in the launch-shape table struct.
No case sets it and nothing reads it — dead field left over from merging the two
implementation lanes. `golangci-lint` does not flag it and it changes no
behaviour. Worth a one-line sweep whenever this file is next touched; not worth
another review cycle.

## Acceptance evidence for the commit-owning mover

This reviewer run supplies no `commit_ack`. The tree is uncommitted per standing
orders on branch `task/TASK-260822-27bvo4-symlink-launcher-507162` (base
`6a9b201`), 4 modified files + 2 new test files, 95 insertions, and matches the
attached `TASK-260822-27bvo4_symlink-launcher.patch`. Note for whoever lands it:
`internal/godriver` does not exist on the main checkout branch
(`handoff/cocoaskills-parity-20260731`); it exists on `main`, still carrying the
unfixed `filepath.EvalSymlinks`.
